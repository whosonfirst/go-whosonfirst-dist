package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-placetypes/filter"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	mode := flag.String("mode", "repo", "Where to read data (to create metafiles) from. If empty then the code will assume the current working directory.")
	out := flag.String("out", "", "Where to store metafiles. If empty then assume metafile are created in the current working directory.")

	// blah blah blah interface definition mismatch atomicfile blah blah blah...
	// stdout := flag.Bool("stdout", false, "Write meta file(s) to STDOUT")

	limit := flag.Int("open-filehandles", 512, "The maximum number of file handles to keep open at any given moment.")

	str_placetypes := flag.String("placetypes", "", "A comma-separated list of placetypes that meta files will be created for. All other placetypes will be ignored.")
	str_roles := flag.String("roles", "", "Role-based filters are not supported yet.")
	str_exclude := flag.String("exclude", "", "A comma-separated list of placetypes that meta files will not be created for.")

	timings := flag.Bool("timings", false, "...")

	procs := flag.Int("processes", 0, "The number of concurrent processes to use. THIS FLAG HAS BEEN DEPRECATED")

	flag.Parse()

	if *procs != 0 {
		log.Println("the -procs flag has been deprecated and will be ignored")
	}

	placetypes := make([]string, 0)
	roles := make([]string, 0)
	exclude := make([]string, 0)

	if *str_placetypes != "" {
		placetypes = strings.Split(*str_placetypes, ",")
	}

	if *str_roles != "" {
		roles = strings.Split(*str_roles, ",")
	}

	if *str_exclude != "" {
		exclude = strings.Split(*str_exclude, ",")
	}

	var abs_root string

	if *out == "" {

		cwd, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		abs_root = cwd
	} else {

		abs_out, err := filepath.Abs(*out)

		if err != nil {
			log.Fatal(err)
		}

		abs_root = abs_out
	}

	placetype_filter, err := filter.NewPlacetypesFilter(placetypes, roles, exclude)

	if err != nil {
		log.Fatal(err)
	}

	mu := new(sync.Mutex)

	throttle := make(chan bool, *limit)

	for i := 0; i < *limit; i++ {
		throttle <- true
	}

	var count int32
	var open int32
	var pending int32
	var scheduled int32

	count = 0
	open = 0
	pending = 0
	scheduled = 0

	filehandles := make(map[string]*atomicfile.File)
	writers := make(map[string]*csv.DictWriter)

	wg := new(sync.WaitGroup)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		atomic.AddInt32(&pending, 1)
		// log.Printf("pending %d scheduled %d\n", pending, scheduled)

		<-throttle

		atomic.AddInt32(&pending, -1)
		atomic.AddInt32(&scheduled, 1)

		wg.Add(1)

		defer func() {
			throttle <- true
			wg.Done()
		}()

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		atomic.AddInt32(&open, 1)
		defer atomic.AddInt32(&open, -1)

		placetype := f.Placetype()

		allow, err := placetype_filter.AllowFromString(placetype)

		if err != nil {
			log.Println(fmt.Sprintf("Unable to validate placetype (%s) for %s", placetype, path))
			return err
		}

		if !allow {
			return nil
		}

		// THIS NEEDS TO BE MORE SOPHISTICATED THAN "just placetype" BUT
		// TODAY IT IS NOT... (20180531/thisisaaronland)

		target := placetype

		row, err := meta.FeatureToRow(f.Bytes())

		if err != nil {
			return err
		}

		r, err := repo.NewDataRepoFromString(whosonfirst.Repo(f))

		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		writer, ok := writers[target]

		if !ok {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			sort.Strings(fieldnames)

			// HOW DOES THIS SQUARE WITH target ABOVE?
			// (20180531/thisisaaronland)

			opts := repo.DefaultFilenameOptions()
			opts.Placetype = placetype

			fname := r.MetaFilename(opts)

			abs_meta := abs_root

			// this is just for backwards compatibility
			// (20180531/thisisaaronland)

			if *out == "" && *mode == "repo" {
				abs_meta = filepath.Join(abs_root, "meta")
			}

			outfile := filepath.Join(abs_meta, fname)

			fh, err := atomicfile.New(outfile, os.FileMode(0644))

			if err != nil {
				return err
			}

			writer, err = csv.NewDictWriter(fh, fieldnames)

			if err != nil {
				return err
			}

			writer.WriteHeader()

			filehandles[placetype] = fh
			writers[placetype] = writer
		}

		writer.WriteRow(row)

		atomic.AddInt32(&count, 1)
		return nil
	}

	i, err := index.NewIndexer(*mode, cb)

	if err != nil {
		log.Fatal(err)
	}

	var index_err error

	t1 := time.Now()

	for _, path := range flag.Args() {

		ta := time.Now()

		err := i.IndexPath(path)

		tb := time.Since(ta)

		if *timings {
			log.Printf("time to prepare %s %v\n", path, tb)
		}

		if err != nil {
			index_err = err
			break
		}
	}

	t2 := time.Since(t1)

	if *timings {
		c := atomic.LoadInt32(&count)
		log.Printf("time to prepare all %d records %v\n", c, t2)
	}

	for _, fh := range filehandles {

		if index_err != nil {
			fh.Abort()
		} else {
			fh.Close()
		}
	}

	if index_err != nil {
		log.Fatal(index_err)
	}
}
