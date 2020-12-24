package main

import (
	_ "github.com/whosonfirst/go-reader-http"
)

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-dist/build"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-dist/utils"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	_ "log"
	"os"
	"path/filepath"
)

func main() {

	opts := options.NewBuildOptions()

	build_sqlite := flag.Bool("build-sqlite", opts.SQLiteCommon, "Build a (common) SQLite distribution for a repo. This flag is DEPRECATED.")
	build_sqlite_common := flag.Bool("build-sqlite-common", opts.SQLiteCommon, "Build a SQLite distribution for a repo, with common tables.")
	build_sqlite_rtree := flag.Bool("build-sqlite-rtree", opts.SQLiteRTree, "Build a SQLite distribution for a repo, with rtree-related tables.")
	build_sqlite_search := flag.Bool("build-sqlite-search", opts.SQLiteSearch, "Build a (common) SQLite distribution for a repo, with search-tables.")
	build_sqlite_all := flag.Bool("build-sqlite-all", false, "Build a SQLite distribution for a repo, with all tables defined by the other -build-sqlite flags.")

	build_meta := flag.Bool("build-meta", opts.Meta, "Build meta files for a repo")
	build_bundle := flag.Bool("build-bundle", opts.Bundle, "Build a bundle distribution for a repo.")

	compress_sqlite := flag.Bool("compress-sqlite", opts.CompressSQLite, "...")
	compress_meta := flag.Bool("compress-meta", opts.CompressMeta, "...")
	compress_bundle := flag.Bool("compress-bundle", opts.CompressBundle, "...")
	compress_all := flag.Bool("compress-all", true, "...")

	compress_max_cpus := flag.Int("compress-max-cpus", opts.CompressMaxCPUs, "Number of concurrent processes to use when compressing distribution items.")

	preserve_checkout := flag.Bool("preserve-checkout", opts.PreserveCheckout, "Do not remove repo from disk after the build process is complete. This is automatically set to true if the -local-checkout flag is true.")
	preserve_sqlite := flag.Bool("preserve-sqlite", opts.PreserveSQLite, "...")
	preserve_meta := flag.Bool("preserve-meta", opts.PreserveMeta, "...")
	preserve_bundle := flag.Bool("preserve-bundle", opts.PreserveBundle, "...")
	preserve_all := flag.Bool("preserve-all", false, "...")

	clone := flag.String("git-clone", opts.Cloner, "Indicate how to clone a repo, using either a native Git binary or the go-git implementation. Currently only the native Git binary is supported.")
	proto := flag.String("git-protocol", opts.Protocol, "Fetch repos using this protocol")
	source := flag.String("git-source", opts.Source, "Fetch repos from this endpoint")
	org := flag.String("git-organization", opts.Organization, "Fetch repos from the user (or organization)")

	local_checkout := flag.Bool("local-checkout", opts.LocalCheckout, "Do not fetch a repo from a remote source but instead use a local checkout on disk")
	local_sqlite := flag.Bool("local-sqlite", opts.LocalSQLite, "Do not build a new SQLite database but use a pre-existing database on disk (this expects to find the database at the same path it would be stored if the database were created from scratch)")

	custom_repo := flag.Bool("custom-repo", true, "Allow custom repo names")

	index_alt_files := flag.Bool("index-alt-files", opts.IndexAltFiles, "Index alternate geometry files.")

	index_relations := flag.Bool("index-relations", false, "Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.")
	relations_uri := flag.String("index-relations-reader-uri", "", "A valid go-reader.Reader URI from which to read data for a relations candidate.")

	combined := flag.Bool("combined", opts.Combined, "Create a single combined distribution from multiple repos.")
	combined_name := flag.String("combined-name", opts.CombinedName, "Distribution name for a single combined distribution from multiple repos.")

	strict := flag.Bool("strict", opts.Strict, "...")
	timings := flag.Bool("timings", opts.Timings, "Display timings during the build process")
	verbose := flag.Bool("verbose", false, "Be chatty")

	workdir := flag.String("workdir", opts.Workdir, "Where to store temporary and final build files. If empty the code will attempt to use the current working directory.")

	flag.Parse()

	logger := log.SimpleWOFLogger()

	if *verbose {
		stdout := io.Writer(os.Stdout)
		logger.AddLogger(stdout, "status")
	}

	// https://github.com/whosonfirst/go-whosonfirst-geojson-v2/issues/5
	// https://github.com/whosonfirst/go-whosonfirst-dist/issues/15

	if *index_alt_files && (*build_meta || *build_bundle) {
		logger.Warning("-index-alt-files is set but please be aware that support for indexing alternate geometries is NOT available for meta files or bundles at this time.")
	}

	if *combined && *combined_name == "" {
		logger.Fatal("Missing -combined-name flag")
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

	if *index_relations {

		ctx := context.Background()
		r, err := reader.NewReader(ctx, *relations_uri)

		if err != nil {
			logger.Fatal("Unable to create go-reader.Reader (%s), %v", *relations_uri, err)
		}

		opts.SQLiteIndexRelations = true
		opts.SQLiteIndexRelationsReader = r
	}

	if *build_sqlite {
		logger.Info("-build-sqlite flag is DEPRECATED but is set so auto-enabling -build-sqlite-common")
		*build_sqlite_common = true
	}

	if *build_sqlite_all {
		logger.Info("-build-sqlite-all flag is set so auto-enabling -build-sqlite-common -build-sqlite-rtree -build-sqlite-search")
		*build_sqlite_common = true
		*build_sqlite_rtree = true
		*build_sqlite_search = true
	}

	opts.Logger = logger

	opts.Cloner = *clone
	opts.Protocol = *proto
	opts.Source = *source
	opts.Organization = *org

	opts.Workdir = *workdir

	opts.SQLiteCommon = *build_sqlite_common
	opts.SQLiteRTree = *build_sqlite_rtree
	opts.SQLiteSearch = *build_sqlite_search

	opts.Meta = *build_meta
	opts.Bundle = *build_bundle

	opts.Combined = *combined
	opts.CombinedName = *combined_name

	opts.IndexAltFiles = *index_alt_files

	opts.LocalCheckout = *local_checkout
	opts.LocalSQLite = *local_sqlite

	opts.CompressSQLite = *compress_sqlite
	opts.CompressMeta = *compress_meta
	opts.CompressBundle = *compress_bundle
	opts.CompressMaxCPUs = *compress_max_cpus

	opts.PreserveCheckout = *preserve_checkout
	opts.PreserveSQLite = *preserve_sqlite
	opts.PreserveMeta = *preserve_meta
	opts.PreserveBundle = *preserve_bundle

	opts.CustomRepo = *custom_repo

	opts.Strict = *strict
	opts.Timings = *timings

	repos := make([]repo.Repo, 0)

	for _, repo_name := range flag.Args() {

		r, err := utils.NewRepo(repo_name, opts)

		if err != nil {
			logger.Fatal("Failed to parse repo '%s', because %s", repo_name, err)
		}

		repos = append(repos, r)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*

		why isn't this triggering a fatal error?
		(20181120/thisisaronland)

		17:12:56.057646 [wof-dist-build] STATUS time to index geojson (797797) : 1m44.498260642s
		17:12:56.057658 [wof-dist-build] STATUS time to index spr (797797) : 13m8.030653553s
		17:12:56.057671 [wof-dist-build] STATUS time to index all (797797) : 49m0.145189371s
		error: Failed to parse tag
		17:13:07.565161 [wof-dist-build] STATUS local sqlite is /usr/local/data/dist/whosonfirst-data-latest.db
		17:47:19.212743 [wof-dist-build] STATUS time to build UNCOMPRESSED distributions for whosonfirst-data 2h12m22.305771409s

	*/

	distribution_items, err := build.BuildDistributionsForRepos(ctx, opts, repos...)

	if err != nil {
		logger.Fatal("Failed to build distributions because %s", err)
	}

	logger.Status("ITEMS %v", distribution_items)

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

		_, err = fh.Write(pretty.Pretty(b))

		if err != nil {
			logger.Warning("Failed to write %s, %s", path, err)
			continue
		}

		fh.Close()

		logger.Status("Wrote inventory %s", path)
	}

	os.Exit(0)
}
