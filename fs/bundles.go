package fs

import (
	"context"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	golog "log"
	"path/filepath"
	"strings"
)

// PLEASE MAKE ME RETURN A LIST OF distribution.Item THINGIES... (20180611/thisisaaronland)

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]string, error) {

	golog.Println("BUNDLE FROM", source)
	golog.Println("BUNDLE WITH", metafiles)

	/*

		2018/07/10 14:28:09 BUNDLE WITH [planet region macroregion country microhood continent ocean marinearea borough dependency disputed neighbourhood locality macrocounty timezone county macrohood campus localadmin empire]
		2018/07/10 14:28:09 HELLO planet
		2018/07/10 14:28:09 HELLO region
		2018/07/10 14:28:09 HELLO macroregion
		2018/07/10 14:28:09 HELLO country
		2018/07/10 14:28:09 HELLO microhood
		2018/07/10 14:28:09 HELLO continent
		2018/07/10 14:28:09 HELLO ocean
		2018/07/10 14:28:09 HELLO marinearea
		2018/07/10 14:28:09 HELLO borough
		2018/07/10 14:28:09 HELLO dependency
		2018/07/10 14:28:09 HELLO disputed
		2018/07/10 14:28:09 HELLO neighbourhood
		2018/07/10 14:28:09 HELLO locality
		2018/07/10 14:28:09 HELLO macrocounty
		2018/07/10 14:28:09 HELLO timezone
		2018/07/10 14:28:09 HELLO county
		2018/07/10 14:28:09 HELLO macrohood
		2018/07/10 14:28:09 HELLO campus
		2018/07/10 14:28:09 HELLO localadmin
		2018/07/10 14:28:09 HELLO empire
		2018/07/10 14:28:09 GO /usr/local/data/dist/empire
		2018/07/10 14:28:09 GO /usr/local/data/dist/empire
		2018/07/10 14:28:09 GO /usr/local/data/dist/empire
		2018/07/10 14:28:09 GO /usr/local/data/dist/empire
		...

	*/

	done_ch := make(chan bool)
	err_ch := make(chan error)
	bundle_ch := make(chan string) // make me an Index thingy, yeah?

	for _, path := range metafiles {

		golog.Println("HELLO", path)

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

				golog.Println("GO", bundle_path)
				err = b.BundleMetafileFromSQLite(ctx, dsn, path)
				golog.Println("WHAT", bundle_path, err)

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
