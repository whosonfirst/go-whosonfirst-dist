package csv

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"strings"
)

type CSVDistributionType struct {
	major string
	minor string
}

func (t *CSVDistributionType) Class() string {
	return "csv"
}

func (t *CSVDistributionType) Major() string {
	return t.major
}

func (t *CSVDistributionType) Minor() string {
	return t.minor
}

func NewCSVDistributionType(major string, minor string) (distribution.DistributionType, error) {

	major = strings.ToLower(major)

	switch major {

	case "meta":
		return NewMetaDistributionType(minor)
	default:
		// pass
	}

	return nil, errors.New("Invalid or unsupported major type")
}

func NewMetaDistributionType(minor string) (distribution.DistributionType, error) {

	minor = strings.ToLower(minor)

	t := CSVDistributionType{
		major: "meta",
		minor: minor,
	}

	return &t, nil
}
