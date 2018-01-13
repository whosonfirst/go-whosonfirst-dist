package main

// THIS IS WET PAINT AND WILL/MIGHT/SHOULD-PROBABLY BE MOVED IN TO ITS OWN
// go-whosonfirst-distributions PACKAGE SO WE CAN REUSE CODE TO BUILD BUNDLES
// AND WHATEVER THE NEXT THING IS (20180112/thisisaaronland)

import (
	"context"
	"flag"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"log"
	"os"
	_ "path/filepath"
	"time"
)

func Build(ctx context.Context, repo string, done_ch chan bool, err_ch chan error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		log.Printf("time to build %s %v\n", repo, t2)
		done_ch <- true
	}()

	select {

	case <-ctx.Done():
		return
	default:
		dir, err := ioutil.TempDir("", repo)

		if err != nil {
			err_ch <- err
			return
		}

		defer func() {
			os.RemoveAll(dir)
		}()

		// MAKE ORG/USER CONFIGURABLE?

		url := fmt.Sprintf("https://github.com/whosonfirst-data/%s.git", repo)

		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL: url,
		})

		if err != nil {
			err_ch <- err
			return
		}
	}

}

func main() {

	flag.Parse()

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repos := flag.Args()
	count := len(repos)

	t1 := time.Now()

	for _, repo := range flag.Args() {
		go Build(ctx, repo, done_ch, err_ch)
	}

	for count > 0 {

		select {
		case <-done_ch:
			count--
		case err := <-err_ch:
			log.Println(err)
			cancel()
		default:
			// pass
		}
	}

	t2 := time.Since(t1)
	log.Printf("time to build all %v\n", t2)
}
