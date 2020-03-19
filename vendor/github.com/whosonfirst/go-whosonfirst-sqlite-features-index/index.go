package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	wof_tables "github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/warning"
	"io"
	"io/ioutil"
	"log"
)

type SQLiteFeaturesLoadRecordFuncOptions struct {
	StrictAltFiles bool
}

func SQLiteFeaturesLoadRecordFunc(opts *SQLiteFeaturesLoadRecordFuncOptions) sql_index.SQLiteIndexerLoadRecordFunc {

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) (interface{}, error) {

		select {

		case <-ctx.Done():
			return nil, nil
		default:

			path, err := wof_index.PathForContext(ctx)

			if err != nil {
				return nil, err
			}

			// we read the whole thing in to memory here so don't
			// have to worry about whether fh is a ReadSeeker if
			// the first attempt to load a primary WOF file fails
			// and we try again for an alt file
			// (20191103/straup)

			body, err := ioutil.ReadAll(fh)

			if err != nil {
				return nil, err
			}

			i, err := feature.NewWOFFeature(body)

			if err != nil && !warning.IsWarning(err) {

				alt, alt_err := feature.NewWOFAltFeature(body)

				if alt_err != nil && !warning.IsWarning(alt_err) {

					msg := fmt.Sprintf("Unable to load %s, because %s (%s)", path, alt_err, err)

					if !opts.StrictAltFiles {
						log.Printf("%s - SKIPPING\n", msg)
						return nil, nil
					}

					return nil, errors.New(msg)
				}

				i = alt
			}

			return i, nil
		}
	}

	return cb
}

func SQLiteFeaturesIndexRelationsFunc(r reader.Reader) sql_index.SQLiteIndexerPostIndexFunc {

	cb := func(ctx context.Context, db sqlite.Database, tables []sqlite.Table, record interface{}) error {

		geojson_t, err := wof_tables.NewGeoJSONTable()

		if err != nil {
			return err
		}

		conn, err := db.Conn()

		if err != nil {
			return err
		}

		f := record.(geojson.Feature)
		body := f.Bytes()

		relations := make(map[int64]bool)
		to_index := make([]geojson.Feature, 0)

		candidates := []string{
			"properties.wof:belongsto",
			"properties.wof:involves",
			"properties.wof:depicts",
		}

		for _, path := range candidates {

			// log.Println("RELATIONS", path)

			rsp := gjson.GetBytes(body, path)

			if !rsp.Exists() {
				// log.Println("MISSING", path)
				continue
			}

			for _, r := range rsp.Array() {
				id := r.Int()
				relations[id] = true
				// log.Println("MATCH", path, id)
			}
		}

		for id, _ := range relations {

			sql := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE id=?", geojson_t.Name())
			row := conn.QueryRow(sql, id)

			var count int
			err = row.Scan(&count)

			if err != nil {
				return err
			}

			if count != 0 {
				continue
			}

			ancestor, err := wof_reader.LoadFeatureFromID(ctx, r, id)

			if err != nil {
				return err
			}

			to_index = append(to_index, ancestor)

			// TO DO: CHECK WHETHER TO INDEX ALT FILES FOR ANCESTOR(S)
			// https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index/issues/3
		}

		for _, record := range to_index {

			for _, t := range tables {

				err = t.IndexRecord(db, record)

				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return cb
}
