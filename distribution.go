package distribution

type Inventory []Item

type Item struct {

        // something something something a private (absolute) path
	// attribute that we can use internally but not expose when
	// this is serialized and published (20180613/thisisaaronland)
	
	Name             string `json:"name"`
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