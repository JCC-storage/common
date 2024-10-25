package cloudstorage

type ObjectStorage struct {
	Manufacturer string `json:"manufacturer"`
	Region       string `json:"region"`
	AK           string `json:"access_key_id"`
	SK           string `json:"secret_access_key"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
}

const (
	HuaweiCloud = "HuaweiCloud"
	AliCloud    = "AliCloud"
	SugonCloud  = "SugonCloud"
)
