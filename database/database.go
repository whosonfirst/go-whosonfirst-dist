package database

import (
	"github.com/whosonfirst/go-whosonfirst-dist"
)

type DatabaseDistributionType struct {
     major string
     minor string
}

func NewDatabaseDistributionType(minor string) (distribution.DistributionType, error) {

     t := DatabaseDistributionType{
       major: "database",
       minor: minor,       	      
     }

     return &t, nil
}