package csv

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-dist/utils"
	meta "github.com/whosonfirst/go-whosonfirst-meta/build"
	meta_options "github.com/whosonfirst/go-whosonfirst-meta/options"
	meta_stats "github.com/whosonfirst/go-whosonfirst-meta/stats"
	_ "log"
	"os"
	"strings"
	"time"
)

type MetaDistribution struct {
	dist.Distribution
	kind       dist.DistributionType
	path       string
	count      int64
	size       int64
	lastupdate int64
}

func (d *MetaDistribution) Type() dist.DistributionType {
	return d.kind
}

func (d *MetaDistribution) Path() string {
	return d.path
}

func (d *MetaDistribution) Count() int64 {
	return d.count
}

func (d *MetaDistribution) Size() int64 {
	return d.size
}

func (d *MetaDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

func (d *MetaDistribution) Compress() (dist.CompressedDistribution, error) {

	path, sha, err := utils.CompressFile(d.path)

	if err != nil {
		return nil, err
	}

	c := MetaCompressedDistribution{
		path: path,
		hash: sha,
	}

	return &c, nil
}

type MetaCompressedDistribution struct {
	path string
	hash string
}

func (c *MetaCompressedDistribution) Path() string {
	return c.path
}

func (c *MetaCompressedDistribution) Hash() string {
	return c.hash
}

func BuildMetaFiles(ctx context.Context, dist_opts *options.BuildOptions, mode string, sources ...string) ([]dist.Distribution, error) {

	meta_opts, err := meta_options.DefaultBuildOptions()

	if err != nil {
		return nil, err
	}

	meta_opts.Workdir = dist_opts.Workdir
	meta_opts.Timings = dist_opts.Timings
	meta_opts.Logger = dist_opts.Logger

	// TBD
	// go-whosonfirst-meta STILL DOESN'T SUPPORT ALT
	// FILES BUT NEEDS TO AND WE NEED TO PASS THE INDEX
	// ALT FILES FLAG (ONCE IT'S BEEN MERGED...)
	// (20190601/thisisaaronland)

	// meta_opts.IndexAltFiles = dist_opts.IndexAltFile

	meta_opts.Combined = dist_opts.Combined
	meta_opts.CombinedName = dist_opts.CombinedName

	metafiles, err := meta.BuildFromIndex(meta_opts, mode, sources)

	if err != nil {
		return nil, err
	}

	// at some point we'll wire all the stats stuff in to meta.BuildFromIndex
	// itself (or equivalent) so we don't have to do a second pass but in the
	// interest of just getting things working we'll suffer to penalty of reading
	// all the CSV files again... (20180715/thisisaaronland)

	done_ch := make(chan bool)
	err_ch := make(chan error)
	stats_ch := make(chan *meta_stats.Stats)

	for _, path := range metafiles {

		go func(path string, stats_ch chan *meta_stats.Stats, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			stats, err := meta_stats.Compile(path)

			if err != nil {
				err_ch <- err
				return
			}

			stats_ch <- stats

		}(path, stats_ch, done_ch, err_ch)
	}

	lookup := make(map[string]*meta_stats.Stats)
	remaining := len(metafiles)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case s := <-stats_ch:
			lookup[s.Path] = s
		case e := <-err_ch:
			return nil, e
		default:
			// pass
		}
	}

	dist_items := make([]dist.Distribution, len(metafiles))

	for i, path := range metafiles {

		stats, ok := lookup[path]

		if !ok {
			return nil, errors.New("Unable to find stats")
		}

		pt := strings.Join(stats.Placetypes, ";")
		k, err := NewMetaDistributionType(pt)

		if err != nil {
			return nil, err
		}

		info, err := os.Stat(path)

		if err != nil {
			return nil, err
		}

		size := info.Size()

		d := MetaDistribution{
			kind:       k,
			path:       path,
			size:       size,
			count:      stats.Count,
			lastupdate: stats.LastUpdate,
		}

		dist_items[i] = &d
	}

	return dist_items, nil
}
