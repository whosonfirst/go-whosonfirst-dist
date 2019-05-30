package options

import (
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"runtime"
)

type BuildOptions struct {
	Cloner       string
	Protocol     string
	Source       string
	Organization string
	Repo         repo.Repo
	Repos        []repo.Repo
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
	CompressMaxCPUs  int
	CustomRepo       bool
	Combined         bool
	CombinedName     string
	Timings          bool
	Strict           bool
	IndexAltFiles    bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	max_cpus := int(float64(runtime.NumCPU()) / 2.0)

	if max_cpus < 1 {
		max_cpus = 1
	}

	opts := BuildOptions{
		Cloner:           "native",
		Source:           "github.com",
		Protocol:         "https",
		Organization:     "whosonfirst-data",
		Repo:             nil,
		Repos:            nil,
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
		CustomRepo:       false,
		CompressSQLite:   true,
		CompressMeta:     true,
		CompressBundle:   true,
		CompressMaxCPUs:  max_cpus,
		Timings:          false,
		Strict:           false,
		Combined:         false,
		CombinedName:     "",
		IndexAltFiles:    false,
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
		PreserveSQLite:   opts.PreserveSQLite,
		PreserveMeta:     opts.PreserveMeta,
		PreserveBundle:   opts.PreserveBundle,
		CompressSQLite:   opts.CompressSQLite,
		CompressMeta:     opts.CompressMeta,
		CompressBundle:   opts.CompressBundle,
		CompressMaxCPUs:  opts.CompressMaxCPUs,
		CustomRepo:       opts.CustomRepo,
		Timings:          opts.Timings,
		Strict:           opts.Strict,
		Combined:         opts.Combined,
		CombinedName:     opts.CombinedName,
		IndexAltFiles:    opts.IndexAltFiles,
	}

	return &clone
}

func DistributionNameFromOptions(opts *BuildOptions) string {

	if opts.Combined {
		return opts.CombinedName
	}

	return opts.Repo.Name()
}

func DistributionRepoFromOptions(opts *BuildOptions) (repo.Repo, error) {

	name := DistributionNameFromOptions(opts)
	return repo.NewCustomRepoFromString(name)
}
