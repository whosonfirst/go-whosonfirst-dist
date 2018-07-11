package fs

import (
	"context"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"path/filepath"
	"strings"
)

// PLEASE MAKE ME RETURN A LIST OF distribution.Item THINGIES... (20180611/thisisaaronland)

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]string, error) {

	done_ch := make(chan bool)
	err_ch := make(chan error)
	bundle_ch := make(chan string) // make me an Index thingy, yeah?

	for _, path := range metafiles {

		go func(dsn string, metafile string, done_ch chan bool, bundle_ch chan string, err_ch chan error) {

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

				err = b.BundleMetafileFromSQLite(ctx, dsn, path)

				if err != nil {
					err_ch <- err
				}

				bundle_ch <- bundle_path
			}

		}(source, path, done_ch, bundle_ch, err_ch)
	}

	bundles := make([]string, 0)
	var build_err error

	remaining := len(metafiles)

	for remaining > 0 {

		select {
		case b := <-bundle_ch:
			bundles = append(bundles, b)
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			build_err = e
			break
		default:
			// pass
		}
	}

	return bundles, build_err
}
