package options

import (
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-repo"
)

type BuildOptions struct {
	Cloner       string
	Protocol     string
	Source       string
	Organization string
	Repo         repo.Repo
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
	LocalSQLite      bool
	PreserveCheckout bool
	PreserveSQLite   bool
	PreserveMeta     bool
	PreserveBundle   bool
	CompressSQLite   bool
	CompressMeta     bool
	CompressBundle   bool

	Timings bool
	Strict  bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Cloner:           "native",
		Source:           "github.com",
		Protocol:         "https",
		Organization:     "whosonfirst-data",
		Repo:             nil,
		SQLite:           true,
		Meta:             false,
		Bundle:           false,
		Workdir:          "",
		Logger:           logger,
		LocalCheckout:    false,
		LocalSQLite:      false,
		PreserveCheckout: false,
		PreserveSQLite:   false,
		PreserveMeta:     false,
		PreserveBundle:   false,
		CompressSQLite:   true,
		CompressMeta:     true,
		CompressBundle:   true,
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
		LocalSQLite:      opts.LocalSQLite,
		PreserveCheckout: opts.PreserveCheckout,
		Timings:          opts.Timings,
		Strict:           opts.Strict,
	}

	return &clone
}
