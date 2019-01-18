package repo

import (
	"fmt"
	"strings"
	"time"
)

type CustomRepo struct {
	Repo
	name string
}

func NewCustomRepoFromString(repo string) (Repo, error) {

	r := CustomRepo{
		name: repo,
	}

	return &r, nil
}

func (r *CustomRepo) String() string {
	return r.Name()
}

func (r *CustomRepo) Name() string {

	return r.name
}

func (r *CustomRepo) MetaFilename(opts *FilenameOptions) string {

	opts.Extension = "csv"
	return r.filename(opts)
}

func (r *CustomRepo) ConcordancesFilename(opts *FilenameOptions) string {

	opts.Suffix = "concordances"
	opts.Extension = "csv"

	return r.filename(opts)
}

func (r *CustomRepo) BundleFilename(opts *FilenameOptions) string {

	opts.Extension = ""
	return r.filename(opts)
}

func (r *CustomRepo) SQLiteFilename(opts *FilenameOptions) string {

	opts.Extension = "db"
	return r.filename(opts)
}

func (r *CustomRepo) filename(opts *FilenameOptions) string {

	parts := []string{
		r.name,
	}

	if opts.Suffix != "" {

		suffix := opts.Suffix

		if opts.Suffix == "{DATED}" {

			now := time.Now()
			suffix = now.Format("20060102")
		}

		parts = append(parts, suffix)
	}

	fname := strings.Join(parts, "-")

	if opts.Extension != "" {
		fname = fmt.Sprintf("%s.%s", fname, opts.Extension)
	}

	return fname
}
