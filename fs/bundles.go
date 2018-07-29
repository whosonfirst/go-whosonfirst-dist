package fs

import (
	"context"
	"github.com/tidwall/gjson"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-dist/utils"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type BundleDistribution struct {
	dist.Distribution
	kind       dist.DistributionType
	path       string
	count      int64
	size       int64
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

func (d *BundleDistribution) Size() int64 {
	return d.size
}

func (d *BundleDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

func (d *BundleDistribution) Compress() (dist.CompressedDistribution, error) {

	path, sha, err := utils.CompressDirectory(d.path)

	if err != nil {
		return nil, err
	}

	c := BundleCompressedDistribution{
		path: path,
		hash: sha,
	}

	return &c, nil
}

type BundleCompressedDistribution struct {
	path string
	hash string
}

func (c *BundleCompressedDistribution) Path() string {
	return c.path
}

func (c *BundleCompressedDistribution) Hash() string {
	return c.hash
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

				f_opts := repo.DefaultFilenameOptions()
				fname := dist_opts.Repo.BundleFilename(f_opts)

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

				// see this? it's massively inefficient - we are double counting
				// everything and we are doing so knowingly for now in the service
				// of a) understanding what kinds of data we need to return for
				// any given distribution and b) because I haven't really figured
				// out how/where/what to store-and-return this data as part of a
				// more generalized indexing process (20180727/thisisaaronland)

				var size int64
				var count int64
				var lastupdate int64

				lastupdate = 0

				mu := new(sync.RWMutex)

				cb := func(path string, info os.FileInfo) error {

					if info.IsDir() {
						return nil
					}

					inc_size := int64(0)
					inc_count := int64(0)
					inc_lastupdate := int64(0)

					inc_size = info.Size()

					if !info.IsDir() {

						// this is salt in the wound having to re-read every file but
						// if we just relied on info.ModTime() we'd get incorrect results
						// since the file itself has just been created (out of the sqlite
						// database) above... (20180727/thisisaaronland)

						fh, err := os.Open(path)

						if err != nil {
							return err
						}

						defer fh.Close()
						
						b, err := ioutil.ReadAll(fh)

						if err != nil {
							return err
						}

						rsp := gjson.GetBytes(b, "properties.wof:lastmodified")

						if rsp.Exists() {
							inc_count = 1
							inc_lastupdate = rsp.Int()
						}
					}

					mu.Lock()
					defer mu.Unlock()

					size += inc_size
					count += inc_count

					lastupdate = int64(math.Max(float64(lastupdate), float64(inc_lastupdate)))

					return nil
				}

				// hrrrrrmmmmm... is there a better way to derive this?
				data_path := filepath.Join(bundle_path, "data")

				cr := crawl.NewCrawler(data_path)
				err = cr.Crawl(cb)

				if err != nil {
					err_ch <- err
					return
				}

				// end of sub-optimal double-counting...

				k, err := NewBundleDistributionType(fname)

				if err != nil {
					err_ch <- err
					return
				}

				d := BundleDistribution{
					kind:       k,
					path:       bundle_path,
					count:      count,
					size:       size,
					lastupdate: lastupdate,
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
