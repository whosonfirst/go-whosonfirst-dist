package git

import (
	"context"
	"log"
	"os/exec"
	"strings"
)

type NativeCloner struct {
	Cloner
	git string
}

func NewNativeCloner() (Cloner, error) {

	// check that git binary is present here...

	cl := NativeCloner{
		git: "git",
	}

	return &cl, nil
}

func (cl *NativeCloner) Clone(ctx context.Context, remote string, local string) error {

	select {

	case <-ctx.Done():
		return nil
	default:

		git_args := []string{
			"clone",
			"--depth",
			"1",
			remote,
			local,
		}

		cmd := exec.Command(cl.git, git_args...)

		log.Println(cl.git, strings.Join(git_args, " "))
		
		_, err := cmd.Output()
		return err
	}
}
