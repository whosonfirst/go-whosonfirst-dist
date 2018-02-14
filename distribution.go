package distribution

type Inventory []Item

type Item struct {
	Name             string `json:"name"`
	NameCompressed   string `json:"name_compressed"`
	Count            int64  `json:"count"`
	Size             int64  `json:"size"`
	SizeCompressed   int64  `json:"size_compressed"`
	Sha256Compressed string `json:"sha256_compressed"`
	LastUpdate       string `json:"last_updated"`
	LastModified     string `json:"lastmodified"`
}
