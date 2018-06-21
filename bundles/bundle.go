package bundles

import (
	"context"
	wof_bundles "github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"path/filepath"
	"strings"
)

func BuildBundle(ctx context.Context, dist_opts *options.BuildOptions, metafiles []string, source string) ([]string, error) {

	bundles := make([]string, 0)

	for _, path := range metafiles {

		// FIX ME: PLEASE USE go-whosonfirst-repo to generate bundle_fname

		bundle_fname := filepath.Base(path)
		bundle_fname = strings.Replace(bundle_fname, ".csv", "", 1)

		bundle_dest := filepath.Join(dist_opts.Workdir, bundle_fname)

		bundle_opts := wof_bundles.DefaultBundleOptions()
		bundle_opts.Destination = bundle_dest
		bundle_opts.Mode = "meta"

		b, err := wof_bundles.NewBundle(bundle_opts)

		if err != nil {
			return bundles, err
		}

		err = b.BundleMetafile(path)

		if err != nil {
			return bundles, err
		}

		bundles = append(bundles, bundle_dest)
	}

	return bundles, nil
}
