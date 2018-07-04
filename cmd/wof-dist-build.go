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

	build_sqlite := flag.Bool("build-sqlite", true, "Build a (common) SQLite distribution for a repo")

	// THESE WILL BE THE NEW-NEW AND THE -build-sqlite FLAG WILL BE DEPRECATED
	// (20180611/thisisaaronland)

	/*
		build_sqlite_common := flag.Bool("build-sqlite-common", true, "...")
		build_sqlite_spatial := flag.Bool("build-sqlite-spatial", true, "...")
		build_sqlite_search := flag.Bool("build-sqlite-search", true, "...")
		build_sqlite_all := flag.Bool("build-sqlite-all", true, "...")
	*/

	build_meta := flag.Bool("build-meta", true, "Build meta files for a repo")
	build_bundle := flag.Bool("build-bundle", false, "Build a bundle distribution for a repo (this flag is enabled but will fail because the code hasn't been implemented)")
	// build_shapefile := flag.Bool("build-shapefile", true, "...")

	compress_sqlite := flag.Bool("compress-sqlite", true, "...")
	compress_meta := flag.Bool("compress-meta", true, "...")
	compress_bundle := flag.Bool("compress-bundle", true, "...")
	compress_all := flag.Bool("compress-all", true, "...")

	preserve_checkout := flag.Bool("preserve-checkout", false, "Do not remove repo from disk after the build process is complete. This is automatically set to true if the -local-checkout flag is true.")
	preserve_sqlite := flag.Bool("preserve-sqlite", false, "...")
	preserve_meta := flag.Bool("preserve-meta", false, "...")
	preserve_bundle := flag.Bool("preserve-bundle", false, "...")
	preserve_all := flag.Bool("preserve-all", false, "...")

	clone := flag.String("git-clone", "native", "Indicate how to clone a repo, using either a native Git binary or the go-git implementation")
	proto := flag.String("git-protocol", "https", "Fetch repos using this protocol")
	source := flag.String("git-source", "github.com", "Fetch repos from this endpoint")
	org := flag.String("git-organization", "whosonfirst-data", "Fetch repos from the user (or organization)")

	local_checkout := flag.Bool("local-checkout", false, "Do not fetch a repo from a remote source but instead use a local checkout on disk")
	local_sqlite := flag.Bool("local-sqlite", false, "Do not build a new SQLite database but use a pre-existing database on disk (this expects to find the database at the same path it would be stored if the database were created from scratch)")

	strict := flag.Bool("strict", false, "...")
	timings := flag.Bool("timings", false, "Display timings during the build process")
	verbose := flag.Bool("verbose", false, "Be chatty")

	workdir := flag.String("workdir", "", "Where to store temporary and final build files. If empty the code will attempt to use the current working directory.")

	flag.Parse()

	if *build_bundle {
		*build_meta = true
	}

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "status")
	}

	if *compress_all {
		*compress_sqlite = true
		*compress_meta = true
		*compress_bundle = true
	}

	if *preserve_all {
		*preserve_checkout = true
		*preserve_sqlite = true
		*preserve_meta = true
		*preserve_bundle = true
	}

	if *local_checkout == true {
		*preserve_checkout = true
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

	opts.Workdir = *workdir

	opts.SQLite = *build_sqlite
	opts.Meta = *build_meta
	opts.Bundle = *build_bundle

	opts.LocalCheckout = *local_checkout
	opts.LocalSQLite = *local_sqlite

	opts.CompressSQLite = *compress_sqlite
	opts.CompressMeta = *compress_meta
	opts.CompressBundle = *compress_bundle

	opts.PreserveCheckout = *preserve_checkout
	opts.PreserveSQLite = *preserve_sqlite
	opts.PreserveMeta = *preserve_meta
	opts.PreserveBundle = *preserve_bundle

	opts.Strict = *strict
	opts.Timings = *timings

	repos := flag.Args()

	err = build.BuildDistributions(opts, repos)

	if err != nil {
		logger.Fatal("Failed to build distributions because %s", err)
	}

	os.Exit(0)
}
