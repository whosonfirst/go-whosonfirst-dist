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
	// these are the new new and will replace "SQLite"
	// SQLiteCommon     bool
	// SQLiteSpatial    bool
	// SQLiteSearch     bool
	Meta             bool
	Bundle           bool
	Workdir          string
	Logger           *log.WOFLogger
	LocalCheckout    bool
	PreserveCheckout bool
	Timings          bool
	Strict           bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Cloner:           "native",
		Source:           "github.com",
		Protocol:         "https",
		Organization:     "whosonfirst-data",
		Repo:             "whosonfirst-data",
		SQLite:           true,
		Meta:             false,
		Bundle:           false,
		Workdir:          "",
		Logger:           logger,
		LocalCheckout:    false,
		PreserveCheckout: false,
		Timings:          false,
		Strict:           false,
	}

	return &opts
}

func (opts *BuildOptions) Clone() *BuildOptions {

	clone := BuildOptions{
		Cloner:           opts.Cloner,
		Source:           opts.Source,
		Protocol:         opts.Protocol,
		Organization:     opts.Organization,
		Repo:             opts.Repo,
		SQLite:           opts.SQLite,
		Meta:             opts.Meta,
		Bundle:           opts.Bundle,
		Workdir:          opts.Workdir,
		Logger:           opts.Logger,
		LocalCheckout:    opts.LocalCheckout,
		PreserveCheckout: opts.PreserveCheckout,
		Timings:          opts.Timings,
		Strict:           opts.Strict,
	}

	return &clone
}
