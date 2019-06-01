package build

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/csv"
	"github.com/whosonfirst/go-whosonfirst-dist/database"
	"github.com/whosonfirst/go-whosonfirst-dist/fs"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-repo"
	_ "log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// take (n) repos and build (n) or 1 (combined) distributions (represented as dist.Item(s))

func BuildDistributionsForRepos(ctx context.Context, opts *options.BuildOptions, repos ...repo.Repo) (map[string][]*dist.Item, error) {

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build distributions for %d repos %v", len(repos), t2)
		}()
	}

	type BuildItem struct {
		Key   string
		Repos []repo.Repo
		Items []*dist.Item
	}

	build_ch := make(chan BuildItem)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	build_func := func(ctx context.Context, r []repo.Repo) {

		defer func() {
			done_ch <- true
		}()

		local_opts := opts.Clone()
		local_opts.Repos = r

		items, err := BuildDistributions(ctx, local_opts)

		key := options.DistributionNameFromOptions(opts)

		if err != nil {
			err_ch <- err
			return
		}

		b := BuildItem{
			Key:   key,
			Repos: r,
			Items: items,
		}

		build_ch <- b
	}

	if opts.Combined {
		go build_func(ctx, repos)
	} else {

		for _, r := range repos {
			go build_func(ctx, []repo.Repo{r})
		}
	}

	items := make(map[string][]*dist.Item)
	var err error

	remaining := 0

	if opts.Combined {
		remaining = 1
	} else {
		remaining = len(repos)
	}

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining--
		case b := <-build_ch:
			items[b.Key] = b.Items
		case e := <-err_ch:
			err = e
			break
		default:
			// pass
		}
	}

	if err != nil {
		return nil, err
	}

	return items, nil
}

// build all the distributions for  (n) repos and then compress each one

func BuildDistributions(ctx context.Context, opts *options.BuildOptions) ([]*dist.Item, error) {

	distributions := make([]dist.Distribution, 0) // uncompressed and private
	items := make([]*dist.Item, 0)                // compressed and public

	if opts.Timings {

		dist_name := options.DistributionNameFromOptions(opts)
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build COMPRESSED distributions for %s %v", dist_name, t2)
		}()
	}

	defer func() {
		cleanupBuildDistributions(ctx, opts, distributions)
	}()

	// actually building stuff...

	distributions, meta, err := buildDistributionsForRepos(ctx, opts)

	if err != nil {
		opts.Logger.Warning("build (buildDistributionsForRepo) for repo %s failed: %s", opts.Repo, err)
		return nil, err
	}

	item_ch := make(chan *dist.Item)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	count_throttle := opts.CompressMaxCPUs

	throttle_ch := make(chan bool, count_throttle)

	for i := 0; i < count_throttle; i++ {
		throttle_ch <- true
	}

	// actually compressing stuff...

	for _, d := range distributions {

		// this appears to be a problem when compressing both the sqlite and bundles
		// distribution at the same time - specifically out-of-memory errors so we
		// need to do some testing to see how a) running them in sequence would affect
		// the overall processing time b) what's needed to be added or tweaked to
		// generate the bundles (and upload them) after the fact - this wouldn't happen
		// in this code but rather by modifying the logic and flags in go-whosonfirst-update
		// to generate and publish but not cleanup the sqlite distribution first and then
		// to generate and publish the bundles from the (newly created) sqlite distribution
		// rather than a fresh git checkout... TBD (20181204/thisisaaronland)

		go func(ctx context.Context, d dist.Distribution, item_ch chan *dist.Item, throttle_ch chan bool, done_ch chan bool, err_ch chan error) {

			defer func() {
				opts.Logger.Status("All done compressing %s", d.Path())
				done_ch <- true
			}()

			opts.Logger.Status("register function to compress %s", d.Path())

			if opts.Timings {

				t1 := time.Now()

				defer func() {
					t2 := time.Since(t1)
					opts.Logger.Status("time to compress %s %v", d.Path(), t2)
				}()
			}

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			ta := time.Now()

			<-throttle_ch

			if opts.Timings {
				tb := time.Since(ta)
				opts.Logger.Status("time to wait to start compressing %s %v", d.Path(), tb)
			}

			defer func() {
				opts.Logger.Status("All done compressing %s (throttle)", d.Path())
				throttle_ch <- true
			}()

			c, err := d.Compress()

			if err != nil {
				err_ch <- err
				return
			}

			i, err := dist.NewItemFromDistribution(d, c, meta)

			if err != nil {
				err_ch <- err
				return
			}

			item_ch <- i

		}(ctx, d, item_ch, throttle_ch, done_ch, err_ch)
	}

	var item_err error
	remaining := len(distributions)

	for remaining > 0 {

		select {

		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			item_err = e
			break
		case i := <-item_ch:
			items = append(items, i)
		default:
			// pass
		}
	}

	if item_err != nil {
		return nil, item_err
	}

	return items, nil
}

/*

take (n) repos and:
1. grab a fresh checkout/clone
2. build a SQLite database
3. optionally build metafiles or build metafiles if building bundles
4. optionally build bundle folder

return:

*/

func buildDistributionsForRepos(ctx context.Context, opts *options.BuildOptions) ([]dist.Distribution, *dist.MetaData, error) {

	if opts.Timings {

		dist_name := options.DistributionNameFromOptions(opts)
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build UNCOMPRESSED distributions for %s %v", dist_name, t2)
		}()
	}

	distributions := make([]dist.Distribution, 0)

	select {

	case <-ctx.Done():
		return nil, nil, nil
	default:
		// pass
	}

	var local_checkouts []string

	var local_sqlite string
	var local_metafiles []string
	var local_bundlefiles []string

	gt, err := git.NewGitToolFromOptions(opts)

	if err != nil {
		return nil, nil, err
	}

	if opts.LocalCheckout {

		for _, r := range opts.Repos {

			// TBD
			// do we need/want a custom local checkout directory...

			fname := r.Name()
			path := filepath.Join(opts.Workdir, fname)

			abs_path, err := filepath.Abs(path)

			if err != nil {
				return nil, nil, err
			}

			_, err = os.Stat(abs_path)

			if err != nil {
				return nil, nil, err
			}

			local_checkouts = append(local_checkouts, abs_path)
		}

	} else if opts.LocalSQLite {

		return nil, nil, errors.New("PLEASE MAKE ME WORK AGAIN...")

	} else {

		repo_paths, err := git.CloneRepo(ctx, gt, opts)

		if err != nil {
			return nil, nil, err
		}

		local_checkouts = repo_paths
	}

	if !opts.PreserveCheckout {
		defer removeLocalCheckouts(opts, local_checkouts)
	}

	opts.Logger.Status("local_checkouts are %s", local_checkouts)

	commit_hashes, err := gt.CommitHashes(local_checkouts...)

	if err != nil {
		opts.Logger.Warning("failed to determine commit hash for %s, %s", local_checkouts, err)
		commit_hashes = make(map[string]string)
	} else {
		opts.Logger.Status("commit hashes are %s (%s)", commit_hashes, local_checkouts)
	}

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		if opts.LocalSQLite {

			return nil, nil, errors.New("Please make me work again")

		} else {

			d, err := database.BuildSQLite(ctx, opts, local_checkouts...)

			if err != nil {
				opts.Logger.Warning("Failed to build SQLite %s because %s", local_sqlite, err)
				return nil, nil, err
			}

			opts.Logger.Status("Built %s without any reported errors", local_sqlite)

			distributions = append(distributions, d)
			local_sqlite = d.Path()
		}

		opts.Logger.Status("local sqlite is %s", local_sqlite)
	}

	_, err = os.Stat(local_sqlite)

	if err != nil {
		return nil, nil, err
	}

	// eventually we should be able to do these two operations in parallel
	// assuming that a SQLite database has been created - and the bundles
	// code has been updated to read from those databases...
	// (20180602/thisisaaronland)

	if opts.Meta {

		mode := "repo"
		sources := local_checkouts

		if opts.SQLite {
			mode = "sqlite"
			sources = []string{local_sqlite}
		}

		opts.Logger.Status("build metafile from %s (%s)", mode, sources)

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		ta := time.Now()

		d_many, err := csv.BuildMetaFiles(ctx, opts, mode, sources...)

		tb := time.Since(ta)

		if err != nil {
			return nil, nil, err
		}

		if len(d_many) == 0 {
			return nil, nil, err
		}

		for _, d := range d_many {
			distributions = append(distributions, d)
			local_metafiles = append(local_metafiles, d.Path())
		}

		opts.Logger.Status("time to build metafiles (%s) %v", strings.Join(local_metafiles, ","), tb)
	}

	if opts.Bundle {

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		ta := time.Now()

		// see notes in fs/bundles.go about meta/bundle filenames (20180731/thisisaaronland)

		bundle_dist, err := fs.BuildBundle(ctx, opts, local_metafiles, local_sqlite)

		tb := time.Since(ta)
		opts.Logger.Status("time to build bundles (%s) %v", strings.Join(local_bundlefiles, ","), tb)

		if err != nil {
			return nil, nil, err
		}

		if len(bundle_dist) == 0 {
			return nil, nil, errors.New("No bundles produced")
		}

		for _, d := range bundle_dist {
			distributions = append(distributions, d)
			local_bundlefiles = append(local_bundlefiles, d.Path())
		}
	}

	dist_name := options.DistributionNameFromOptions(opts)

	meta := &dist.MetaData{
		CommitHashes: commit_hashes,
		Repo:         dist_name,
	}

	return distributions, meta, nil
}

func cleanupBuildDistributions(ctx context.Context, opts *options.BuildOptions, distributions []dist.Distribution) error {

	if opts.Timings {

		dist_name := options.DistributionNameFromOptions(opts)
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to remove uncompressed files for %s %v", dist_name, t2)
		}()
	}

	rm := func(path string) {

		opts.Logger.Status("remove uncompressed file %s", path)

		info, err := os.Stat(path)

		if os.IsNotExist(err) {
			return
		}

		if err != nil {
			opts.Logger.Warning("Failed to stat path '%s' : %s", path, err)
			return
		}

		if info.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			opts.Logger.Warning("Failed to remove '%s' : %s", path, err)
		}
	}

	wg := new(sync.WaitGroup)

	for _, d := range distributions {

		t := d.Type()

		switch t.Class() {

		case "csv":

			if t.Major() == "meta" && opts.PreserveMeta {
				continue
			}

		case "database":

			if t.Major() == "sqlite" && opts.PreserveSQLite {
				continue
			}

		case "fs":

			if t.Major() == "bundle" && opts.PreserveBundle {
				continue
			}

		default:
			// pass
		}

		wg.Add(1)

		go func(path string, wg *sync.WaitGroup) {

			defer wg.Done()
			rm(path)

		}(d.Path(), wg)
	}

	wg.Wait()

	return nil
}

func removeLocalCheckouts(opts *options.BuildOptions, local_checkouts []string) error {

	for _, path_checkout := range local_checkouts {

		err := os.RemoveAll(path_checkout)

		if err != nil {
			opts.Logger.Status("failed to remove %s, %s", path_checkout, err)
		}
	}

	return nil
}
