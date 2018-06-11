package csv

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	meta "github.com/whosonfirst/go-whosonfirst-meta/build"
	meta_options "github.com/whosonfirst/go-whosonfirst-meta/options"
)

// Not really sure about this signature... (20180604/thisisaaronland)

// PLEASE MAKE ME RETURN A LIST OF distribution.Item THINGIES... (20180611/thisisaaronland)

func BuildMetaFiles(ctx context.Context, dist_opts *options.BuildOptions, mode string, source string) ([]string, error) {

	meta_opts, err := meta_options.DefaultBuildOptions()

	if err != nil {
		return nil, err
	}

	meta_opts.Workdir = dist_opts.Workdir
	meta_opts.Timings = dist_opts.Timings
	meta_opts.Logger = dist_opts.Logger

	return meta.BuildFromIndex(meta_opts, mode, []string{source})
}
