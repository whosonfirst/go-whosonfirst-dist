package database

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"strings"
)

type DatabaseDistributionType struct {
	major string
	minor string
}

func (t *DatabaseDistributionType) Class() string {
	return "database"
}

func (t *DatabaseDistributionType) Major() string {
	return t.major
}

func (t *DatabaseDistributionType) Minor() string {
	return t.minor
}

func NewDatabaseDistributionType(major string, minor string) (distribution.DistributionType, error) {

	major = strings.ToLower(major)

	switch major {

	case "sqlite":
		return NewSQLiteDistributionType(minor)
	default:
		// pass
	}

	return nil, errors.New("Invalid or unsupported major type")
}

func NewSQLiteDistributionType(minor string) (distribution.DistributionType, error) {

	minor = strings.ToLower(minor)

	switch minor {

	case "common":
		// pass
	default:
		return nil, errors.New("Invalid or unsupported minor type")
	}

	t := DatabaseDistributionType{
		major: "sqlite",
		minor: minor,
	}

	return &t, nil
}
