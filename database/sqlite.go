package database

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io/ioutil"
	_ "log"
	"path/filepath"
	"time"
)

func BuildSQLite(ctx context.Context, local_repo string, opts *options.BuildOptions) (string, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, local_repo, opts)
}

func BuildSQLiteCommon(ctx context.Context, local_repo string, opts *options.BuildOptions) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		if opts.Timings {
			t1 := time.Now()

			defer func() {
				t2 := time.Since(t1)
				opts.Logger.Info("time to generate (common) sqlite tables %v", t2)
			}()
		}

		name := filepath.Base(local_repo)

		dir := fmt.Sprintf("%s-sqlite", name)
		root, err := ioutil.TempDir("", dir) // PLEASE MAKE THIS CONFIGURABLE

		if err != nil {
			return "", err
		}

		fname := fmt.Sprintf("%s-latest.db", name)
		dsn := filepath.Join(root, fname)

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

		idx.Timings = opts.Timings
		idx.Logger = opts.Logger

		err = idx.IndexPaths("repo", []string{local_repo})

		if err != nil {
			return "", err
		}

		return dsn, nil
	}
}