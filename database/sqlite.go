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
	"os"
	"path/filepath"
	"time"
)

type SQLiteDistribution {
     distribution.Distribution
     type string	// distribution.DatabaseDistributionType
     path string
     count int64
     lastupdate int64
}

func (d *SQLiteDistribution) Type() string {
     return d.type 
}

func (d *SQLiteDistribution) Path() string {
     return d.path
}

func (d *SQLiteDistribution) Count() int64 {
     return d.count
}

// PLEASE MAKE ME RETURN A distribution.Item thingy... (20180611/thisisaaronland)

func BuildSQLite(ctx context.Context, local_repo string, opts *options.BuildOptions) (*distribution.Item, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, local_repo, opts)
}

// PLEASE MAKE ME RETURN A distribution.Item thingy... (20180611/thisisaaronland)

func BuildSQLiteCommon(ctx context.Context, local_repo string, opts *options.BuildOptions) (*distribution.Item, error) {

     
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

		info, err := os.Stat(dsn)

		if err != nil {
		   return nil, err
		}

		fsize := info.Size()
		lastmod := info.ModTime()
		
		item := distribution.Item {
			// things we need want
			// path: dsn,
			// major: "sqlite",
			// minor: "common",
			Name: fname,
			NameCompressed: "",
			Count: count,
			Size: fsize,
 			SizeCompressed: 0,
			Sha256Compressed: "",
			LastUpdate: lastupdate,
			LastModified: lastmod,
			Repo: opts.Repo,
			Commit: "",
		}

		// compress stuff here or later? if we do it here then by the time we
		// call NewDistributionItemFromDB - see notes after this - then we'll
		// have all the stuff we need to build a distribution.Item thingy...
		// (20180613/thisisaaronland)

		// ideally we could just return NewDistributionItemFromDB(dsn) but I am
		// not sure about open/closed database handles - maybe we just don't care
		// and assume the function is private and pass it a laundry list of things
		// to do whatever we need... (20180613/thisisaaronland)

		return &item, nil
	}
}

func NewDistributionItemFromDB(path string) (*distribution.Item, error) {

	info, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	size := info.Size()
	tm_lastmod := info.ModTime()

	db, err := database.NewDBWithDriver("sqlite3", path)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	conn, err := db.Conn()

	if err != nil {
		return nil, err
	}

	t, err := tables.NewGeoJSONTable()

	if err != nil {
		return nil, err
	}

	var lastupdate int64
	var count int64

	sql_lastupdate := fmt.Sprintf("SELECT MAX(lastmodified) FROM %s", t.Name())
	sql_count := fmt.Sprintf("SELECT COUNT(id) FROM %s", t.Name())

	row_lastupdate := conn.QueryRow(sql_lastupdate)
	err = row_lastupdate.Scan(&lastupdate)

	if err != nil {
		return nil, err
	}

	tm_lastupdate := time.Unix(lastupdate, 0)

	row_count := conn.QueryRow(sql_count)
	err = row_count.Scan(&count)

	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)

	i := distribution.Item{
		Name:             name,
		Count:            count,
		Size:             size,
		LastModified:     tm_lastmod.Format(time.RFC3339),
		LastUpdate:       tm_lastupdate.Format(time.RFC3339),
		NameCompressed:   "XX",
		Sha256Compressed: "XX",
		SizeCompressed:   -1,
	}

	return &i, nil
}
