package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist/build"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	_ "log"
	"os"
	"path/filepath"
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

	build_meta := flag.Bool("build-meta", false, "Build meta files for a repo")
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

	// PLEASE MAKE ME WORK, YEAH... (20180704/thisisaaronland)
	// remote_sqlite := flag.Bool("remote-sqlite", false, "Do not build a new SQLite database but use a pre-existing database that is stored (on dist.whosonfirst.org for now)")

	strict := flag.Bool("strict", false, "...")
	timings := flag.Bool("timings", false, "Display timings during the build process")
	verbose := flag.Bool("verbose", false, "Be chatty")

	workdir := flag.String("workdir", "", "Where to store temporary and final build files. If empty the code will attempt to use the current working directory.")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "status")
	}

	if *compress_all {

		logger.Info("-compress-all flag is set so auto-enabling -compress-sqlite -compress-meta -compress-bundle")

		*compress_sqlite = true
		*compress_meta = true
		*compress_bundle = true
	}

	if *preserve_all {

		logger.Info("-preserve-all flag is set so auto-enabling -preserve-checkout -preserve-sqlite -preserve-meta -preserve-bundle")

		*preserve_checkout = true
		*preserve_sqlite = true
		*preserve_meta = true
		*preserve_bundle = true
	}

	if *local_checkout == true {

		logger.Info("local-checkout flag is set so auto-enabling -preserve-checkout")

		*preserve_checkout = true
	}

	if *local_sqlite == true {

		logger.Info("-local-sqlite flag is set so auto-enabling -preserve-sqlite")

		*preserve_sqlite = true
	}

	if *build_bundle {

		logger.Info("-build-bundle flag is set so auto-enabling -build-meta")

		*build_meta = true
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

	repos := make([]repo.Repo, 0)

	for _, repo_name := range flag.Args() {

		var r repo.Repo
		var err error

		if opts.LocalCheckout {
			r, err = repo.NewDataRepoFromPath(repo_name)
		} else {
			r, err = repo.NewDataRepoFromString(repo_name)
		}

		if err != nil {
			logger.Fatal("Failed to parse repo '%s', because %s", repo_name, err)
		}

		repos = append(repos, r)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	distribution_items, err := build.BuildDistributionsForRepos(ctx, opts, repos...)

	if err != nil {
		logger.Fatal("Failed to build distributions because %s", err)
	}

	for repo_name, items := range distribution_items {

		fname := fmt.Sprintf("%s-inventory.json", repo_name)
		path := filepath.Join(opts.Workdir, fname)

		b, err := json.Marshal(items)

		if err != nil {
			logger.Warning("Failed to encode %s, %s", path, err)
			continue
		}

		fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			logger.Warning("Failed to open %s for writing, %s", path, err)
			continue
		}

		_, err = fh.Write(b)

		if err != nil {
			logger.Warning("Failed to write %s, %s", path, err)
			continue
		}

		fh.Close()

		logger.Status("Wrote inventory %s", path)
	}

	os.Exit(0)
}
