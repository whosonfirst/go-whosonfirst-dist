package git

import (
	"context"
	gogit "gopkg.in/src-d/go-git.v4"
	"os"
)

type GolangCloner struct {
	Cloner
}

func NewGolangCloner() (Cloner, error) {

	cl := GolangCloner{}

	return &cl, nil
}

func (cl *GolangCloner) Clone(ctx context.Context, remote string, local string) error {

	select {

	case <-ctx.Done():
		return nil
	default:

		_, err := gogit.PlainCloneContext(ctx, local, false, &gogit.CloneOptions{
			URL:      remote,
			Depth:    1,
			Progress: os.Stdout,
		})

		return err
	}
}
