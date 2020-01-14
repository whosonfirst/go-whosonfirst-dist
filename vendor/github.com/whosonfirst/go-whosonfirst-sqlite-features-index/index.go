package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/warning"
	"io"
	"io/ioutil"
	"log"
)

type SQLiteFeaturesIndexerCallbackOptions struct {
	StrictAltFiles bool
}

func DefaultSQLiteFeaturesIndexerCallbackOptions() *SQLiteFeaturesIndexerCallbackOptions {

	opts := &SQLiteFeaturesIndexerCallbackOptions{
		StrictAltFiles: true,
	}

	return opts
}

func SQLiteFeaturesIndexerCallback(opts *SQLiteFeaturesIndexerCallbackOptions) sql_index.SQLiteIndexerFunc {

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

func NewDefaultSQLiteFeaturesIndexer(db sqlite.Database, to_index []sqlite.Table) (*sql_index.SQLiteIndexer, error) {

	opts := DefaultSQLiteFeaturesIndexerCallbackOptions()
	cb := SQLiteFeaturesIndexerCallback(opts)

	return NewSQLiteFeaturesIndexerWithCallback(db, to_index, cb)
}

func NewSQLiteFeaturesIndexerWithCallback(db sqlite.Database, to_index []sqlite.Table, cb sql_index.SQLiteIndexerFunc) (*sql_index.SQLiteIndexer, error) {
	return sql_index.NewSQLiteIndexer(db, to_index, cb)
}
