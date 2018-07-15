package csv

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	meta "github.com/whosonfirst/go-whosonfirst-meta/build"
	meta_options "github.com/whosonfirst/go-whosonfirst-meta/options"
	meta_stats "github.com/whosonfirst/go-whosonfirst-meta/stats"
	_ "log"
	"strings"
	"time"
)

type MetaDistribution struct {
	distribution.Distribution
	kind       distribution.DistributionType
	path       string
	count      int64
	lastupdate int64
}

func (d *MetaDistribution) Type() distribution.DistributionType {
	return d.kind
}

func (d *MetaDistribution) Path() string {
	return d.path
}

func (d *MetaDistribution) Count() int64 {
	return d.count
}

func (d *MetaDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

// func BuildMetaFiles(ctx context.Context, dist_opts *options.BuildOptions, d distribution.Distribution) ([]distribution.Distribution, error) {

func BuildMetaFiles(ctx context.Context, dist_opts *options.BuildOptions, mode string, source string) ([]distribution.Distribution, error) {

	meta_opts, err := meta_options.DefaultBuildOptions()

	if err != nil {
		return nil, err
	}

	meta_opts.Workdir = dist_opts.Workdir
	meta_opts.Timings = dist_opts.Timings
	meta_opts.Logger = dist_opts.Logger

	// mode := "sqlite"	// d.Mode() ?
	// source := []string{ d.Path() }

	metafiles, err := meta.BuildFromIndex(meta_opts, mode, []string{source})

	if err != nil {
		return nil, err
	}

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

	dist_items := make([]distribution.Distribution, len(metafiles))

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

		d := MetaDistribution{
			kind:       k,
			path:       path,
			count:      stats.Count,
			lastupdate: stats.LastUpdate,
		}

		dist_items[i] = &d
	}

	return dist_items, nil
}
