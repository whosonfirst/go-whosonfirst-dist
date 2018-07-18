package fs

import (
	"context"
	"errors"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	meta_stats "github.com/whosonfirst/go-whosonfirst-meta/stats"
	"path/filepath"
	"strings"
	"time"
)

type BundleDistribution struct {
	dist.Distribution
	kind       dist.DistributionType
	path       string
	count      int64
	lastupdate int64
}

func (d *BundleDistribution) Type() dist.DistributionType {
	return d.kind
}

func (d *BundleDistribution) Path() string {
	return d.path
}

func (d *BundleDistribution) Count() int64 {
	return d.count
}

func (d *BundleDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

func (d *BundleDistribution) Compress() (dist.CompressedDistribution, error) {
	return nil, errors.New("Please write me")
}

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]dist.Distribution, error) {

	done_ch := make(chan bool)
	err_ch := make(chan error)
	dist_ch := make(chan dist.Distribution)

	for _, path := range metafiles {

		go func(dsn string, metafile string, dist_ch chan dist.Distribution, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			select {

			case <-ctx.Done():
				return
			default:

				abs_path, err := filepath.Abs(metafile)

				if err != nil {
					err_ch <- err
					return
				}

				// please make this less bad... (20180710/thisisaaronland)

				fname := filepath.Base(abs_path)
				ext := filepath.Ext(fname)
				fname = strings.Replace(fname, ext, "", -1)

				bundle_path := filepath.Join(dist_opts.Workdir, fname)

				bundle_opts := wof_bundles.DefaultBundleOptions()
				bundle_opts.Mode = "sqlite"
				bundle_opts.Destination = bundle_path
				bundle_opts.Logger = dist_opts.Logger

				b, err := wof_bundles.NewBundle(bundle_opts)

				if err != nil {
					err_ch <- err
					return
				}

				err = b.BundleMetafileFromSQLite(ctx, dsn, abs_path)

				if err != nil {
					err_ch <- err
					return
				}

				// see this... yeah - it's not yet clear to me where to
				// to generate stats in the various DoThisFromThat packages
				// or how to return them so we're just (re) crunching meta
				// files for now... (20180717/thisisaaronland)

				stats, err := meta_stats.Compile(abs_path)

				if err != nil {
					err_ch <- err
					return
				}

				k, err := NewBundleDistributionType(fname)

				if err != nil {
					err_ch <- err
					return
				}

				d := BundleDistribution{
					kind:       k,
					path:       bundle_path,
					count:      stats.Count,
					lastupdate: stats.LastUpdate,
				}

				dist_ch <- &d
			}

		}(source, path, dist_ch, done_ch, err_ch)
	}

	build_items := make([]dist.Distribution, 0)
	var build_err error

	remaining := len(metafiles)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case d := <-dist_ch:
			build_items = append(build_items, d)
		case e := <-err_ch:
			build_err = e
			break
		default:
			// pass
		}
	}

	return build_items, build_err
}
