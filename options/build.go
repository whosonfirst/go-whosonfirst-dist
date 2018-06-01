package options

import (
	"github.com/whosonfirst/go-whosonfirst-log"
)

type BuildOptions struct {
     	Cloner	  string
	Protocol  string
	Source	  string
	Organization string
	Repo         string
	SQLite       bool
	Bundle       bool
	WorkDir      string
	Logger       *log.WOFLogger
	Local        bool
}

func NewBuildOptions() *BuildOptions {

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Cloner: "native",
		Source: "github.com",
		Protocol: "https",
		Organization: "whosonfirst-data",
		Repo:         "whosonfirst-data",
		SQLite:       true,
		Bundle:       false,
		WorkDir:      "",
		Logger:       logger,
		Local:        false,
	}

	return &opts
}
