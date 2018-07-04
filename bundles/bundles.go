package bundles

import (
	"context"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
)

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]string, error) {

	bundles := make([]string, 0)

	bundle_opts := wof_bundles.DefaultBundleOptions()
	bundle_opts.Mode = "sqlite"
	bundle_opts.Destination = dist_opts.Workdir
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

	// FIX ME - UPDATE bundles HERE...

	return bundles, nil
}
