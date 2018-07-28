package main

// this is work in progress - you should assume that anything and
// everything might change still (20180728/thisisaaronland)

import (
	"bytes"
	// "context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type PublishOptions struct {
	Workdir string
}

func PublishInventory(inv *dist.Inventory, opts *PublishOptions) error {

	t1 := time.Now()

	defer func() {
		log.Println("time to publish inventory %v", time.Since(t1))
	}()

	wg := new(sync.WaitGroup)

	for _, item := range *inv {

		wg.Add(1)

		go func() {

			defer wg.Done()
			err := PublishItem(item, opts)

			if err != nil {
				log.Printf("Failed to publish %s %s\n", item.Name, err)
			}

		}()
	}

	wg.Wait()

	return nil
}

/*

  {
      "name": "whosonfirst-data-constituency-us-latest.csv",
      "type": "x-urn:whosonfirst:csv:meta#constituency",
      "name_compressed": "whosonfirst-data-constituency-us-latest.csv.bz2",
      "count": 7184,
      "size": 2227902,
      "size_compressed": 524455,
      "sha256_compressed": "9cff892bb4a5317a1bcad1c88755b74e0a8415134e5c0de41cc14507594c0eb1",
      "last_updated": "2018-07-24T15:05:33Z",
      "lastmodified": "2018-07-28T15:46:22Z"
   }

*/

func PublishItem(item *dist.Item, opts *PublishOptions) error {

	n := item.NameCompressed
	t := item.Type

	suffix := fmt.Sprintf("-%d.", item.LastUpdate)
	n_ts := strings.Replace(n, "-latest.", suffix, -1)

	// what is NewDistributionTypeFromString(t) ...

	t = strings.Replace(t, "x-urn:whosonfirst:", "", -1)

	var prefix string

	// this will all be made less-shit...

	if strings.HasPrefix(t, "csv:meta") {
		prefix = "bundles"
	} else if strings.HasPrefix(t, "database:sqlite") {
		prefix = "sqlite"
	} else if strings.HasPrefix(t, "fs:bundle") {
		prefix = "bundles"
	} else {
		return errors.New("Invalid or unsupported prefix")
	}

	source := filepath.Join(opts.Workdir, n)

	dest_ts := filepath.Join(prefix, n_ts)
	dest_latest := filepath.Join(prefix, n)

	inv_ts := fmt.Sprintf(dest_ts, ".json")
	inv_latest := fmt.Sprintf(dest_latest, ".json")

	i, err := json.Marshal(item)

	if err != nil {
		return err
	}

	i = pretty.Pretty(i)

	err = publishFile(source, dest_ts, opts)

	if err != nil {
		return err
	}

	err = publishBytes(i, inv_ts, opts)

	if err != nil {
		return err
	}

	err = publishFile(source, dest_latest, opts)

	if err != nil {
		return err
	}

	err = publishBytes(i, inv_latest, opts)

	if err != nil {
		return err
	}

	return nil
}

func publishFile(source string, dest string, opts *PublishOptions) error {

	fh, err := os.Open(source)

	if err != nil {
		return err
	}

	defer fh.Close()

	return publish(fh, dest, opts)
}

func publishBytes(b []byte, dest string, opts *PublishOptions) error {

	r := bytes.NewReader(b)
	fh := ioutil.NopCloser(r)

	return publish(fh, dest, opts)
}

func publish(r io.ReadCloser, dest string, opts *PublishOptions) error {

	log.Println("publish to", dest)
	return nil
}

func main() {

	workdir := flag.String("workdir", "", "Where to store temporary and final build files. If empty the code will attempt to use the current working directory.")

	flag.Parse()

	opts := &PublishOptions{
		Workdir: *workdir,
	}

	for _, repo_name := range flag.Args() {

		r, err := repo.NewDataRepoFromString(repo_name)

		if err != nil {
			log.Fatal(err)
		}

		// PLEASE FIX ME... this should be in a library...

		fname := fmt.Sprintf("%s-inventory.json", r.Name())
		path := filepath.Join(*workdir, fname)

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		defer fh.Close()

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		var inv *dist.Inventory

		err = json.Unmarshal(body, &inv)

		if err != nil {
			log.Fatal(err)
		}

		// ctx, cancel := context.WithCancel(context.Background())
		// defer cancel()

		PublishInventory(inv, opts)
	}
}
