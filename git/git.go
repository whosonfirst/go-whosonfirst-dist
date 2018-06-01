package git

import (
	"context"
	"errors"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"io/ioutil"
	"time"
)

type Cloner interface {
	Clone(context.Context, string, string) error
}

func NewClonerFromOptions(opts *options.BuildOptions) (Cloner, error) {

	var cl Cloner
	var err error

	switch opts.Cloner {

	case "native":
		cl, err = NewNativeCloner()
	case "golang":
		cl, err = NewGolangCloner()
	default:
		err = errors.New("Invalid cloner")
	}

	return cl, err
}

func CloneRepo(ctx context.Context, opts *options.BuildOptions) (string, error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
	}()

	uri := "{protocol}://{source}/{organization}/{repo}.git"

	template, err := uritemplates.Parse(uri)

	if err != nil {
		return "", err
	}

	values := make(map[string]interface{})

	values["protocol"] = opts.Protocol
	values["source"] = opts.Source
	values["organization"] = opts.Organization
	values["repo"] = opts.Repo

	remote, err := template.Expand(values)

	if err != nil {
		return "", err
	}

	// MAKE THIS CONFIGURABLE

	local, err := ioutil.TempDir("", opts.Repo)

	if err != nil {
		return "", err
	}

	cl, err := NewClonerFromOptions(opts)

	// SOMETHING SOMETHING LFS

	err = cl.Clone(ctx, remote, local)
	return local, err
}
