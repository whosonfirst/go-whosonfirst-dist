package fs

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"strings"
)

type FSDistributionType struct {
	major string
	minor string
}

func (t *FSDistributionType) Class() string {
	return "fs"
}

func (t *FSDistributionType) Major() string {
	return t.major
}

func (t *FSDistributionType) Minor() string {
	return t.minor
}

func NewFSDistributionType(major string, minor string) (distribution.DistributionType, error) {

	major = strings.ToLower(major)

	switch major {

	case "bundle":
		return NewBundleDistributionType(minor)
	default:
		// pass
	}

	return nil, errors.New("Invalid or unsupported major type")
}

func NewBundleDistributionType(minor string) (distribution.DistributionType, error) {

	minor = strings.ToLower(minor)

	t := FSDistributionType{
		major: "bundle",
		minor: minor,
	}

	return &t, nil
}
