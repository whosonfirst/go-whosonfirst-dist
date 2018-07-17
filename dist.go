package dist

import (
       "fmt"
       "os"
       "time"
)

// for internal use

type Distribution interface {
     Type() DistributionType
     Path() string
     Count() int64
     LastUpdate() time.Time
}

type DistributionType interface {
     Class() string
     Major() string
     Minor() string
}

// for external publication

type Inventory []Item

type Item struct {
	Name             string `json:"name"`
	Type		 string `json:"type"`
	NameCompressed   string `json:"name_compressed"`
	Count            int64  `json:"count"`
	Size             int64  `json:"size"`
	SizeCompressed   int64  `json:"size_compressed"`
	Sha256Compressed string `json:"sha256_compressed"`
	LastUpdate       string `json:"last_updated"`
	LastModified     string `json:"lastmodified"`
	Repo		 string `json:"repo,omitempty"`
	Commit		 string `json:"commit,omitempty"`
}

func (i *Item) String() string {
     return i.Name
}

func NewItemFromDistribution(d Distribution) (*Item, error) {

     info, err := os.Stat(d.Path())

     if err != nil{
     	return nil, err
     }

     fname := info.Name()
     fsize := info.Size()
     lastmod := info.ModTime()

     count := d.Count()
     lastupdate := d.LastUpdate()

     str_lastmod := lastmod.Format(time.RFC3339)
     str_lastupdate := lastupdate.Format(time.RFC3339)

     t := d.Type()
     
     str_type := fmt.Sprintf("x-urn:whosonfirst:%s:%s#%s", t.Class(), t.Major(), t.Minor())
     
     i := Item {
     	Name: fname,
	Type: str_type,
	
	Size: fsize,
	Count: count,

	LastUpdate: str_lastupdate,
	LastModified: str_lastmod,
	
	Repo: "",
	Commit: "",
	
	NameCompressed: "",
	SizeCompressed: -1,
	Sha256Compressed: "",
     }

     return &i, nil
}