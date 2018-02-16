package main

// THIS IS WET PAINT AND WILL/MIGHT/SHOULD-PROBABLY BE MOVED IN TO ITS OWN
// go-whosonfirst-distributions PACKAGE SO WE CAN REUSE CODE TO BUILD BUNDLES
// AND WHATEVER THE NEXT THING IS (20180112/thisisaaronland)

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"gopkg.in/src-d/go-git.v4"
	"io"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"time"
)

type BuildOptions struct {
	Organization string
	Repo         string
	SQLite       bool
	Bundle       bool
	WorkDir      string
	Logger       *log.WOFLogger
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Organization: "whosonfirst-data",
		Repo:         "whosonfirst-data",
		SQLite:       true,
		Bundle:       false,
		WorkDir:      "",
		Logger:       logger,
	}

	return &opts
}

func Build(ctx context.Context, opts *BuildOptions, done_ch chan bool, err_ch chan error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		opts.Logger.Status("time to build %s %v\n", opts.Repo, t2)
		done_ch <- true
	}()

	var local_repo string

	select {

	case <-ctx.Done():
		return
	default:

		repo, err := Clone(ctx, opts)

		if err != nil {
			err_ch <- err
			return
		}

		defer func() {
			os.RemoveAll(repo)
		}()

		local_repo = repo
	}

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			dsn, err := BuildSQLite(ctx, opts, local_repo)

			if err != nil {
				err_ch <- err
				return
			}

			opts.Logger.Status("CREATED %s", dsn)
		}
	}
}

func BuildSQLite(ctx context.Context, opts *BuildOptions, local_repo string) (string, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, opts, local_repo)
}

func BuildSQLiteCommon(ctx context.Context, opts *BuildOptions, local_repo string) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build sqlite db for %s %v\n", opts.Repo, t2)
		}()

		dir := fmt.Sprintf("%s-sqlite", opts.Repo)
		root, err := ioutil.TempDir("", dir)

		if err != nil {
			return "", err
		}

		fname := fmt.Sprintf("%s-latest.db", opts.Repo)
		dsn := filepath.Join(root, fname)

		db, err := database.NewDBWithDriver("sqlite3", dsn)

		if err != nil {
			return "", err
		}

		defer db.Close()

		err = db.LiveHardDieFast()

		if err != nil {
			return "", err
		}

		to_index, err := tables.CommonTablesWithDatabase(db)

		if err != nil {
			return "", err
		}

		idx, err := index.NewDefaultSQLiteFeaturesIndexer(db, to_index)

		if err != nil {
			return "", err
		}

		idx.Timings = true
		idx.Logger = opts.Logger

		err = idx.IndexPaths("repo", []string{local_repo})

		if err != nil {
			return "", err
		}

		return dsn, nil
	}
}

func Clone(ctx context.Context, opts *BuildOptions) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
		}()

		// MAKE THIS CONFIGURABLE

		dir, err := ioutil.TempDir("", opts.Repo)

		if err != nil {
			return "", err
		}

		// DO NOT HOG-TIE THIS TO GITHUB...

		url := fmt.Sprintf("https://github.com/%s/%s.git", opts.Organization, opts.Repo)

		// SOMETHING SOMETHING SOMETHING LFS...

		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL: url,
		})

		return dir, err
	}
}

func main() {

	flag.Parse()

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repos := flag.Args()
	count := len(repos)

	t1 := time.Now()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	for _, repo := range flag.Args() {

		opts := NewBuildOptions()
		opts.Logger = logger
		opts.Repo = repo

		go Build(ctx, opts, done_ch, err_ch)
	}

	for count > 0 {

		select {
		case <-done_ch:
			count--
		case err := <-err_ch:
			logger.Error("%s", err)
			cancel()
		default:
			// pass
		}
	}

	t2 := time.Since(t1)
	logger.Status("time to build all %v\n", t2)
}
