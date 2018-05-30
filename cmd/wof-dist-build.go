package main

// THIS IS WET PAINT AND WILL/MIGHT/SHOULD-PROBABLY BE MOVED IN TO ITS OWN
// go-whosonfirst-distributions PACKAGE SO WE CAN REUSE CODE TO BUILD BUNDLES
// AND WHATEVER THE NEXT THING IS (20180112/thisisaaronland)

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"time"
)

type BuildOptions struct {
	Organization string
	Repo         string
	SQLite       bool
	Bundle       bool
	WorkDir      string
	Logger       *log.WOFLogger
	Local        bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Organization: "whosonfirst-data",
		Repo:         "whosonfirst-data",
		SQLite:       true,
		Bundle:       false,
		WorkDir:      "",
		Logger:       logger,
		Local:        false,
	}

	return &opts
}

func Build(ctx context.Context, opts *BuildOptions, done_ch chan bool, err_ch chan error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		opts.Logger.Status("time to build %s %v\n", opts.Repo, t2)
		done_ch <- true
	}()

	var local_repo string

	select {

	case <-ctx.Done():
		return
	default:

		if !opts.Local {

			clone_opts := git.CloneOptions{
				Logger:       opts.Logger,
				Repo:         opts.Repo,
				Organization: opts.Organization,
			}

			repo, err := git.Clone(ctx, &clone_opts)

			if err != nil {
				err_ch <- err
				return
			}

			// make me a flag or something (20180405/thisisaaronland)

			defer func() {
				os.RemoveAll(repo)
			}()

			local_repo = repo

		} else {
			local_repo = opts.Repo
		}

	}

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			dsn, err := BuildSQLite(ctx, opts, local_repo)

			if err != nil {
				err_ch <- err
				return
			}

			opts.Logger.Status("CREATED %s", dsn)
		}
	}
}

func BuildSQLite(ctx context.Context, opts *BuildOptions, local_repo string) (string, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, opts, local_repo)
}

func BuildSQLiteCommon(ctx context.Context, opts *BuildOptions, local_repo string) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build sqlite db for %s %v\n", opts.Repo, t2)
		}()

		name := opts.Repo

		if opts.Local {
			name = filepath.Base(name)
		}

		dir := fmt.Sprintf("%s-sqlite", name)
		root, err := ioutil.TempDir("", dir)

		if err != nil {
			return "", err
		}

		fname := fmt.Sprintf("%s-latest.db", name)
		dsn := filepath.Join(root, fname)

		opts.Logger.Status("DSN %s", dsn)

		db, err := database.NewDBWithDriver("sqlite3", dsn)

		if err != nil {
			return "", err
		}

		defer db.Close()

		err = db.LiveHardDieFast()

		if err != nil {
			return "", err
		}

		to_index, err := tables.CommonTablesWithDatabase(db)

		if err != nil {
			return "", err
		}

		idx, err := index.NewDefaultSQLiteFeaturesIndexer(db, to_index)

		if err != nil {
			return "", err
		}

		idx.Timings = true
		idx.Logger = opts.Logger

		err = idx.IndexPaths("repo", []string{local_repo})

		if err != nil {
			return "", err
		}

		return dsn, nil
	}
}

func main() {

	local := flag.Bool("local", false, "...")

	flag.Parse()

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repos := flag.Args()
	count := len(repos)

	t1 := time.Now()

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	for _, repo := range flag.Args() {

		opts := NewBuildOptions()
		opts.Logger = logger
		opts.Repo = repo
		opts.Local = *local

		go Build(ctx, opts, done_ch, err_ch)
	}

	for count > 0 {

		select {
		case <-done_ch:
			count--
		case err := <-err_ch:
			logger.Error("%s", err)
			cancel()
		default:
			// pass
		}
	}

	t2 := time.Since(t1)
	logger.Status("time to build all %v\n", t2)
}
