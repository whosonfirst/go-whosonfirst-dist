package main

// THIS IS WET PAINT AND WILL/MIGHT/SHOULD-PROBABLY BE MOVED IN TO ITS OWN
// go-whosonfirst-distributions PACKAGE SO WE CAN REUSE CODE TO BUILD BUNDLES
// AND WHATEVER THE NEXT THING IS (20180112/thisisaaronland)

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/sqlite"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	_ "log"
	"os"
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

			// cl, _ := git.NewNativeCloner()
			cl, _ := git.NewGolangCloner()			

			clone_opts := git.CloneOptions{
				Cloner:       cl,
				Logger:       opts.Logger,
				Repo:         opts.Repo,
				Organization: opts.Organization,
			}

			repo, err := git.CloneRepo(ctx, &clone_opts)

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

	opts.Logger.Status("LOCAL %s", local_repo)

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			dsn, err := sqlite.BuildSQLite(ctx, local_repo)

			if err != nil {
				err_ch <- err
				return
			}

			opts.Logger.Status("CREATED %s", dsn)
		}
	}


}

func main() {

	local := flag.Bool("local", false, "...")
	build_sqlite := flag.Bool("build-sqlite", true, "...")	

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
		opts.SQLite = *build_sqlite
		
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
