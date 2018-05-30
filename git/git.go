package git

import (
	"github.com/whosonfirst/go-whosonfirst-log"
)

type CloneOptions struct {
	Organization string
	Repo         string
	Logger       *log.WOFLogger
}
