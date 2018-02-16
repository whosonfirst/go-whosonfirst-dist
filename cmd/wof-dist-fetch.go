package main

import (
	"compress/bzip2"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

type RegexpFlag []*regexp.Regexp

func (r *RegexpFlag) String() string {
	return "..."
}

func (r *RegexpFlag) Set(value string) error {

	re, err := regexp.Compile(value)

	if err != nil {
		return err
	}

	*r = append(*r, re)
	return nil
}

func RetrieveInventory(url string) (*distribution.Inventory, error) {

	rsp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	var inv distribution.Inventory

	err = json.Unmarshal(body, &inv)

	if err != nil {
		return nil, err
	}

	return &inv, nil
}

func FetchItem(ctx context.Context, item distribution.Item, source *url.URL, dest string, error_ch chan error, done_ch chan bool) {

	defer func() {
		done_ch <- true
	}()

	select {
	case <-ctx.Done():
		return
	default:

		source.Path = filepath.Join(source.Path, item.NameCompressed)
		remote := source.String()

		local := filepath.Join(dest, item.Name) // note the uncompressed name since we're going to decompress on the fly

		rsp, err := http.Get(remote)

		if err != nil {
			error_ch <- err
			return
		}

		if rsp.StatusCode != 200 {
			err := fmt.Sprintf("%s returned HTTP error %s", remote, rsp.Status)
			error_ch <- errors.New(err)
			return
		}

		defer rsp.Body.Close()

		bz_reader := bzip2.NewReader(rsp.Body)

		fh, err := os.OpenFile(local, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			error_ch <- err
			return
		}

		defer fh.Close()

		b, err := io.Copy(fh, bz_reader)

		if err != nil {
			error_ch <- err
			return
		}

		log.Printf("WROTE %s (%d bytes)\n", local, b)
	}
}

func main() {

	var include RegexpFlag
	var exclude RegexpFlag

	var inventory = flag.String("inventory", "https://dist.whosonfirst.org/sqlite/inventory.json", "The URL of a valid distribution inventory file")
	var dest = flag.String("dest", os.TempDir(), "Where distribution files should be written")

	flag.Var(&include, "include", "A valid regular expression for comparing against an item's filename, for inclusion")
	flag.Var(&exclude, "exclude", "A valid regular expression for comparing against an item's filename, for exclusion")

	flag.Parse()

	source, err := url.Parse(*inventory)

	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(*dest)

	if err != nil {
		log.Fatal(err)
	}

	if !info.IsDir() {
		log.Fatal(errors.New("Destination is not a directory"))
	}

	inv, err := RetrieveInventory(source.String())

	if err != nil {
		log.Fatal(err)
	}

	to_fetch := make([]distribution.Item, 0)

	for _, item := range *inv {

		name := item.Name
		ok := true

		for _, fl := range include {

			if !fl.Match([]byte(name)) {
				ok = false
				break
			}
		}

		for _, fl := range exclude {

			if fl.Match([]byte(name)) {
				ok = false
				break
			}
		}

		if !ok {
			continue
		}

		to_fetch = append(to_fetch, item)
	}

	count := len(to_fetch)

	if count == 0 {
		log.Fatal("Nothing to fetch")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	error_ch := make(chan error)
	done_ch := make(chan bool)

	for _, item := range to_fetch {

		src, err := url.Parse(source.String())

		if err != nil {
			log.Fatal(err)
		}

		src.Path = filepath.Dir(src.Path)
		go FetchItem(ctx, item, src, *dest, error_ch, done_ch)
	}

	for count > 0 {
		select {
		case <-done_ch:
			count -= 1
		case err := <-error_ch:
			log.Println(err)
			cancel()
		default:
			// pass
		}
	}

}
