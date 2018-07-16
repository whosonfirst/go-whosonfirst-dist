package database

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	_ "log"
	"path/filepath"
	"time"
)

type SQLiteDistribution struct {
	dist.Distribution
	kind       dist.DistributionType
	path       string
	count      int64
	lastupdate int64
}

func (d *SQLiteDistribution) Type() dist.DistributionType {
	return d.kind
}

func (d *SQLiteDistribution) Path() string {
	return d.path
}

func (d *SQLiteDistribution) Count() int64 {
	return d.count
}

func (d *SQLiteDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

func BuildSQLite(ctx context.Context, local_repo string, opts *options.BuildOptions) (dist.Distribution, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, local_repo, opts)
}

// PLEASE MAKE ME RETURN A distribution.Item thingy... (20180611/thisisaaronland)

func BuildSQLiteCommon(ctx context.Context, local_repo string, opts *options.BuildOptions) (dist.Distribution, error) {

	select {

	case <-ctx.Done():
		return nil, nil
	default:

		if opts.Timings {

			t1 := time.Now()

			defer func() {
				t2 := time.Since(t1)
				opts.Logger.Info("time to generate (common) sqlite tables %v", t2)
			}()
		}

		// SOMETHING SOMETHING SOMETHING PLEASE USE
		// go-whosonfirst-repo (20180611/thisisaaronland)

		name := filepath.Base(local_repo)

		fname := fmt.Sprintf("%s-latest.db", name)
		dsn := filepath.Join(opts.Workdir, fname)

		db, err := database.NewDBWithDriver("sqlite3", dsn)

		if err != nil {
			return nil, err
		}

		defer db.Close()

		err = db.LiveHardDieFast()

		if err != nil {
			return nil, err
		}

		to_index, err := tables.CommonTablesWithDatabase(db)

		if err != nil {
			return nil, err
		}

		idx, err := index.NewDefaultSQLiteFeaturesIndexer(db, to_index)

		if err != nil {
			return nil, err
		}

		idx.Timings = opts.Timings
		idx.Logger = opts.Logger

		err = idx.IndexPaths("repo", []string{local_repo})

		if err != nil {
			return nil, err
		}

		t, err := tables.NewGeoJSONTable()

		if err != nil {
			return nil, err
		}

		conn, err := db.Conn()

		if err != nil {
			return nil, err
		}

		var count int
		var lastupdate int

		sql := fmt.Sprintf("SELECT COUNT(id) FROM %s", t.Name())
		row := conn.QueryRow(sql)

		err = row.Scan(&count)

		if err != nil {
			return nil, err
		}

		sql = fmt.Sprintf("SELECT MAX(lastmodified) FROM %s", t.Name())
		row = conn.QueryRow(sql)

		err = row.Scan(&lastupdate)

		if err != nil {
			return nil, err
		}

		k, err := NewSQLiteDistributionType("common")

		if err != nil {
			return nil, err
		}

		d := SQLiteDistribution{
			kind:       k,
			path:       dsn,
			count:      int64(count),
			lastupdate: int64(lastupdate),
		}

		return &d, nil
	}
}
