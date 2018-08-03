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
	Size() int64
	LastUpdate() time.Time

	Compress() (CompressedDistribution, error)
}

type DistributionType interface {
	Class() string
	Major() string
	Minor() string
}

type CompressedDistribution interface {
	Path() string
	Hash() string
	// something something something Type() ?
}

// for external publication

type Inventory []*Item

// all of this (the Item stuff) is still wet paint and will almost
// certainly change (20180717/thisisaaronland)

type Item struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	NameCompressed   string `json:"name_compressed"`
	Count            int64  `json:"count"`
	Size             int64  `json:"size"`
	SizeCompressed   int64  `json:"size_compressed"`
	Sha256Compressed string `json:"sha256_compressed"`
	LastUpdate       string `json:"last_updated"`
	LastModified     string `json:"last_modified"`
	Repo             string `json:"repo,omitempty"`
	Commit           string `json:"commit,omitempty"`
}

func (i *Item) String() string {
	return i.Name
}

func NewItemFromDistribution(d Distribution, c CompressedDistribution) (*Item, error) {

	info, err := os.Stat(d.Path())

	if err != nil {
		return nil, err
	}

	info_compressed, err := os.Stat(c.Path())

	if err != nil {
		return nil, err
	}

	fname := info.Name()
	lastmod := info.ModTime()

	fname_compressed := info_compressed.Name()
	fsize_compressed := info_compressed.Size()
	hash_compressed := c.Hash()

	fsize := d.Size()

	count := d.Count()
	lastupdate := d.LastUpdate()

	str_lastmod := lastmod.Format(time.RFC3339)
	str_lastupdate := lastupdate.Format(time.RFC3339)

	t := d.Type()

	str_type := fmt.Sprintf("x-urn:whosonfirst:%s:%s#%s", t.Class(), t.Major(), t.Minor())

	i := Item{
		Name: fname,
		Type: str_type,

		Size:  fsize,
		Count: count,

		LastUpdate:   str_lastupdate,
		LastModified: str_lastmod,

		Repo:   "",
		Commit: "",

		NameCompressed:   fname_compressed,
		SizeCompressed:   fsize_compressed,
		Sha256Compressed: hash_compressed,
	}

	return &i, nil
}
