package main

import (
	"flag"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-placetypes/filter"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	repo_path := flag.String("repo", "", "Where to read data (to create metafiles) from. If empty then the code will assume the current working directory.")
	out := flag.String("out", "", "Where to store metafiles. If empty then assume metafile are created in a child folder of 'repo' called 'meta'.")

	procs := flag.Int("processes", runtime.NumCPU()*2, "The number of concurrent processes to use.")
	limit := flag.Int("open-filehandles", 512, "The maximum number of file handles to keep open at any given moment.")

	str_placetypes := flag.String("placetypes", "", "A comma-separated list of placetypes that meta files will be created for. All other placetypes will be ignored.")
	str_roles := flag.String("roles", "", "Role-based filters are not supported yet.")
	str_exclude := flag.String("exclude", "", "A comma-separated list of placetypes that meta files will not be created for.")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	if *repo_path == "" {

		cwd, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		*repo_path = cwd
	}

	abs_repo, err := filepath.Abs(*repo_path)

	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(abs_repo)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal(fmt.Sprintf("Invalid repo directory (%s)", abs_repo))
	}

	r, err := repo.NewDataRepoFromPath(abs_repo)

	if err != nil {
		log.Fatal(err)
	}

	var abs_meta string

	if *out == "" {
		abs_meta = filepath.Join(abs_repo, "meta")
	} else {
		abs_meta = *out
	}

	info, err = os.Stat(abs_meta)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal(fmt.Sprintf("Invalid meta directory (%s)", abs_meta))
	}

	abs_data := filepath.Join(abs_repo, "data")

	info, err = os.Stat(abs_data)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal(fmt.Sprintf("Invalid data directory (%s)", abs_data))
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

	defer func() {

		for _, fh := range filehandles {
			fh.Close()
		}
	}()

	wg := new(sync.WaitGroup)

	callback := func(path string, info os.FileInfo) error {

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

		if info.IsDir() {
			return nil
		}

		wof, err := uri.IsWOFFile(path)

		if err != nil {
			log.Fatal(err)
		}

		if !wof {
			return nil
		}

		alt, err := uri.IsAltFile(path)

		if err != nil {
			log.Fatal(err)
		}

		if alt {
			return nil
		}

		fh, err := os.Open(path)

		if err != nil {
			msg := fmt.Sprintf("Failed to open %s because %s (%d open filehandles)", path, err, atomic.LoadInt32(&open))
			log.Fatal(msg)
		}

		defer fh.Close()

		atomic.AddInt32(&open, 1)
		defer atomic.AddInt32(&open, -1)

		feature, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		placetype := gjson.GetBytes(feature, "properties.wof:placetype").String()

		allow, err := placetype_filter.AllowFromString(placetype)

		if err != nil {
			log.Println(fmt.Sprintf("Unable to validate placetype (%s) for %s", placetype, path))
			// log.Fatal(err)
		}

		// log.Printf("Allow %s : %t\n", placetype, allow)

		if !allow {
			return nil
		}

		row, err := meta.FeatureToRow(feature)

		if err != nil {
			log.Fatal(err)
		}

		mu.Lock()

		writer, ok := writers[placetype]

		if !ok {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			sort.Strings(fieldnames)

			opts := repo.DefaultFilenameOptions()
			opts.Placetype = placetype

			fname := r.MetaFilename(opts)

			outfile := filepath.Join(abs_meta, fname)

			fh, err := atomicfile.New(outfile, os.FileMode(0644))

			if err != nil {
				log.Fatal(err)
			}

			writer, err = csv.NewDictWriter(fh, fieldnames)
			writer.WriteHeader()

			filehandles[placetype] = fh
			writers[placetype] = writer
		}

		writer.WriteRow(row)

		mu.Unlock()

		atomic.AddInt32(&count, 1)
		return nil
	}

	t1 := time.Now()

	cr := crawl.NewCrawler(abs_data)
	err = cr.Crawl(callback)

	wg.Wait()

	t2 := time.Since(t1)
	log.Printf("time to dump %d features: %v\n", count, t2)

	if err != nil {
		log.Fatal(err)
	}
}
