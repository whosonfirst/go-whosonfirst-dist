package fs

import (
	"context"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
)

// PLEASE MAKE ME RETURN A LIST OF distribution.Item THINGIES... (20180611/thisisaaronland)

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]string, error) {

	bundles := make([]string, 0)

	dest := dist_opts.Workdir	// FIX ME TO INCLUDE BUNDLE NAME OR SOMETHING...
	
	bundle_opts := wof_bundles.DefaultBundleOptions()
	bundle_opts.Mode = "sqlite"
	bundle_opts.Destination = dest
	bundle_opts.Logger = dist_opts.Logger

	b, err := wof_bundles.NewBundle(bundle_opts)

	if err != nil {
		return bundles, err
	}

	// this condition (no metafiles) is unlikely to ever happen
	// but still... (20180704/thisisaaronland)

	if len(metafiles) > 0 {
		err = b.BundleMetafilesFromSQLite(source, metafiles...)
	} else {
		err = b.Bundle(source)
	}

	if err != nil {
		return bundles, err
	}

	bundles = append(bundles, dest)

	return bundles, nil
}
