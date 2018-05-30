package git

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	gogit "gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"time"
)

type CloneOptions struct {
	Organization string
	Repo         string
	Logger       *log.WOFLogger
}

func Clone(ctx context.Context, opts *CloneOptions) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
		}()

		// MAKE THIS CONFIGURABLE

		dir, err := ioutil.TempDir("", opts.Repo)

		if err != nil {
			return "", err
		}

		// DO NOT HOG-TIE THIS TO GITHUB...

		url := fmt.Sprintf("https://github.com/%s/%s.git", opts.Organization, opts.Repo)

		// SOMETHING SOMETHING SOMETHING LFS...

		_, err = gogit.PlainClone(dir, false, &gogit.CloneOptions{
			URL: url,
		})

		return dir, err
	}
}
