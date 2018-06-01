package main

// THIS IS WET PAINT AND WILL/MIGHT/SHOULD-PROBABLY BE MOVED IN TO ITS OWN
// go-whosonfirst-distributions PACKAGE SO WE CAN REUSE CODE TO BUILD BUNDLES
// AND WHATEVER THE NEXT THING IS (20180112/thisisaaronland)

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-dist/build"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	_ "log"
	"os"
)

func main() {

	build_sqlite := flag.Bool("build-sqlite", true, "...")
	// build_meta := flag.Bool("build-meta", true, "...")
	// build_bundle := flag.Bool("build-bundle", true, "...")
	// build_shapefile := flag.Bool("build-shapefile", true, "...")

	clone := flag.String("git-clone", "native", "...")
	proto := flag.String("git-protocol", "https", "...")
	source := flag.String("git-source", "github.com", "...")
	org := flag.String("git-organization", "whosonfirst-data", "...")

	local := flag.Bool("local", false, "...")
	strict := flag.Bool("strict", false, "...")
	timings := flag.Bool("timings", false, "...")
	verbose := flag.Bool("verbose", false, "...")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "status")
	}

	opts := options.NewBuildOptions()
	opts.Logger = logger

	opts.Cloner = *clone
	opts.Protocol = *proto
	opts.Source = *source
	opts.Organization = *org

	opts.SQLite = *build_sqlite

	opts.Local = *local
	opts.Strict = *strict
	opts.Timings = *timings

	repos := flag.Args()

	err := build.BuildDistributions(opts, repos)

	if err != nil {
		logger.Fatal("Failed to build distributions because %s", err)
	}

	os.Exit(0)
}
