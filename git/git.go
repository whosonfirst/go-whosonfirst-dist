package git

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-log"
)

type Cloner interface {
	Clone(context.Context, *CloneOptions) (string, error)
}

type CloneOptions struct {
	Organization string
	Repo         string
	Logger       *log.WOFLogger
}
