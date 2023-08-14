package lyveapi

type Credentials struct {
	AccountId string `json:"accountId"`
	AccessKey string `json:"accessKey"`
	Secret    string `json:"secret"`
}

type CreateServiceAcctReq struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type CreateServiceAcctResp struct {
	Id             string `json:"id"`
	AccessKey      string `json:"accessKey"`
	Secret         string `json:"secret"`
	ExpirationDate string `json:"expirationDate"`
}

// Depending on the API used, some of these fields may or may not be used.
type ServiceAcct struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
	ReadyState  bool     `json:"readyState,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type ServiceAcctUpdateReq struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type ServiceAcctList []ServiceAcct

type Permission struct {
	Id          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Type        PermissionType `json:"type,omitempty"`
	ReadyState  bool           `json:"readyState,omitempty"`
	Actions     Action         `json:"actions,omitempty"`
	Prefix      string         `json:"prefix,omitempty"`
	Buckets     []string       `json:"buckets,omitempty"`
	Policy      string         `json:"policy,omitempty"`
	CreateTime  string         `json:"createTime,omitempty"`
}

func (p *Permission) IsPolicyPermission() bool {
	return p.Type == Policy
}

type PermissionList []Permission

type PermissionType string

const (
	AllBuckets   PermissionType = "all-buckets"
	BucketPrefix PermissionType = "bucket-prefix"
	BucketNames  PermissionType = "bucket-names"
	Policy       PermissionType = "policy"
)

type Action string

const (
	AllOperations Action = "all-operations"
	ReadOnly      Action = "read-only"
	WriteOnly     Action = "write-only"
)

type Token struct {
	Token         string `json:"token,omitempty"`
	ExpirationSec string `json:"expirationSec,omitempty"`
}
