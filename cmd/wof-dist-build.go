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

	/*
	build_sqlite_common := flag.Bool("build-sqlite-common", true, "...")
	build_sqlite_spatial := flag.Bool("build-sqlite-spatial", true, "...")
	build_sqlite_search := flag.Bool("build-sqlite-search", true, "...")
	build_sqlite_all := flag.Bool("build-sqlite-all", true, "...")
	*/
		
	build_meta := flag.Bool("build-meta", true, "...")
	build_bundle := flag.Bool("build-bundle", false, "...")
	// build_shapefile := flag.Bool("build-shapefile", true, "...")

	clone := flag.String("git-clone", "native", "...")
	proto := flag.String("git-protocol", "https", "...")
	source := flag.String("git-source", "github.com", "...")
	org := flag.String("git-organization", "whosonfirst-data", "...")

	local_checkout := flag.Bool("local-checkout", false, "...")
	preserve_checkout := flag.Bool("preserve-checkout", false, "...")

	strict := flag.Bool("strict", false, "...")
	timings := flag.Bool("timings", false, "...")
	verbose := flag.Bool("verbose", false, "...")

	workdir := flag.String("workdir", "", "...")

	flag.Parse()

	if *build_bundle {
		*build_meta = true
	}

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "status")
	}

	if *workdir == "" {

		cwd, err := os.Getwd()

		if err != nil {
			logger.Fatal("Unable to determine current working directory because %s", err)
		}

		*workdir = cwd
	}

	info, err := os.Stat(*workdir)

	if err != nil {
		logger.Fatal("Unable to validate working directory because %s", err)
	}

	if !info.IsDir() {
		logger.Fatal("-workdir is not actually a directory")
	}

	opts := options.NewBuildOptions()
	opts.Logger = logger

	opts.Cloner = *clone
	opts.Protocol = *proto
	opts.Source = *source
	opts.Organization = *org

	opts.SQLite = *build_sqlite
	opts.Meta = *build_meta
	opts.Bundle = *build_bundle
	opts.Workdir = *workdir

	opts.LocalCheckout = *local_checkout
	opts.PreserveCheckout = *preserve_checkout

	opts.Strict = *strict
	opts.Timings = *timings

	repos := flag.Args()

	err = build.BuildDistributions(opts, repos)

	if err != nil {
		logger.Fatal("Failed to build distributions because %s", err)
	}

	os.Exit(0)
}
