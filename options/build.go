package options

import (
	"github.com/whosonfirst/go-whosonfirst-log"
)

type BuildOptions struct {
	Cloner       string
	Protocol     string
	Source       string
	Organization string
	Repo         string
	SQLite       bool
	Bundle       bool
	WorkDir      string
	Logger       *log.WOFLogger
	Local        bool
	Timings      bool
	Strict       bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Cloner:       "native",
		Source:       "github.com",
		Protocol:     "https",
		Organization: "whosonfirst-data",
		Repo:         "whosonfirst-data",
		SQLite:       true,
		Bundle:       false,
		WorkDir:      "",
		Logger:       logger,
		Local:        false,
		Timings:      false,
		Strict:       false,
	}

	return &opts
}

func (opts *BuildOptions) Clone() *BuildOptions {

	clone := BuildOptions{
		Cloner:       opts.Cloner,
		Source:       opts.Source,
		Protocol:     opts.Protocol,
		Organization: opts.Organization,
		Repo:         opts.Repo,
		SQLite:       opts.SQLite,
		Bundle:       opts.Bundle,
		WorkDir:      opts.WorkDir,
		Logger:       opts.Logger,
		Local:        opts.Local,
		Timings:      opts.Timings,
		Strict:       opts.Strict,
	}

	return &clone
}
