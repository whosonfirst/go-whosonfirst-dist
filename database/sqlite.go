package database

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-dist/utils"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"os"
	"path/filepath"
	"time"
)

type SQLiteDistribution struct {
	dist.Distribution
	kind       dist.DistributionType
	path       string
	count      int64
	size       int64
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

func (d *SQLiteDistribution) Size() int64 {
	return d.size
}

func (d *SQLiteDistribution) LastUpdate() time.Time {
	return time.Unix(d.lastupdate, 0)
}

func (d *SQLiteDistribution) Compress() (dist.CompressedDistribution, error) {

	path, sha, err := utils.CompressFile(d.path)

	if err != nil {
		return nil, err
	}

	c := SQLiteCompressedDistribution{
		path: path,
		hash: sha,
	}

	return &c, nil
}

type SQLiteCompressedDistribution struct {
	path string
	hash string
}

func (c *SQLiteCompressedDistribution) Path() string {
	return c.path
}

func (c *SQLiteCompressedDistribution) Hash() string {
	return c.hash
}

func BuildSQLite(ctx context.Context, opts *options.BuildOptions, local_repos ...string) (dist.Distribution, error) {

	// ADD HOOKS FOR -spatial and -search databases... (20180216/thisisaaronland)
	return BuildSQLiteCommon(ctx, opts, local_repos...)
}

func BuildSQLiteCommon(ctx context.Context, opts *options.BuildOptions, local_repos ...string) (dist.Distribution, error) {

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Info("time to generate (common) sqlite tables %v", t2)
		}()
	}

	dsn_repo, err := options.DistributionRepoFromOptions(opts)

	if err != nil {
		return nil, err
	}

	dsn_opts := repo.DefaultFilenameOptions()
	dsn_fname := dsn_repo.SQLiteFilename(dsn_opts)

	dsn := filepath.Join(opts.Workdir, dsn_fname)

	db, err := database.NewDBWithDriver("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	err = db.LiveHardDieFast()

	if err != nil {
		return nil, err
	}

	geojson_opts, err := tables.DefaultGeoJSONTableOptions()

	if err != nil {
		return nil, err
	}

	geojson_opts.IndexAltFiles = opts.IndexAltFiles

	table_opts := &tables.CommonTablesOptions{
		GeoJSON: geojson_opts,
	}

	to_index, err := tables.CommonTablesWithDatabaseAndOptions(db, table_opts)

	if err != nil {
		return nil, err
	}

	record_opts := &index.SQLiteFeaturesLoadRecordFuncOptions{
		StrictAltFiles: true,
	}

	record_func := index.SQLiteFeaturesLoadRecordFunc(record_opts)

	idx_opts := &sql_index.SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	if opts.SQLiteIndexRelations {
		relations_func := index.SQLiteFeaturesIndexRelationsFunc(opts.SQLiteIndexRelationsReader)
		idx_opts.PostIndexFunc = relations_func
	}

	idx, err := sql_index.NewSQLiteIndexer(idx_opts)

	idx.Timings = opts.Timings
	idx.Logger = opts.Logger

	err = idx.IndexPaths("repo", local_repos)

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

	info, err := os.Stat(dsn)

	if err != nil {
		return nil, err
	}

	size := info.Size()

	d := SQLiteDistribution{
		kind:       k,
		path:       dsn,
		count:      int64(count),
		size:       size,
		lastupdate: int64(lastupdate),
	}

	return &d, nil
}
