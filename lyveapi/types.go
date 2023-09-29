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

// Bucket describes usage of a particular bucket.
type Bucket struct {
	Name    string  `json:"name"`    // Name is the name of the given bucket
	UsageGB float64 `json:"usageGB"` // UsageGB reports GBs used by bucket
}

// UsageInBytes converts from the gigabytes reported by the API to bytes. A
// gigabyte (GB) is 1e9 bytes.
func (b Bucket) UsageInBytes() float64 {
	return 1e9 * b.UsageGB
}

// SubAccount contains summary usage information for the given sub-account.
type SubAccount struct {
	// SubAccountName is the human-friendly name of the given sub-account
	SubAccountName string `json:"subAccountName"`
	// SubAccountId is the unique identifier of the given sub-account
	SubAccountId string `json:"subAccountId"`
	// CreateTime is the timestamp of this bucket's creation
	CreateTime string `json:"createTime"`
	// UsageGB is the amount of GBs consumed by the given sub-account
	UsageGB float64 `json:"usageGB,omitempty"`
	// Users is the number of users tied to the given sub-account
	Users int `json:"users,omitempty"`
	// ServiceAccounts is the number of service accounts tied to the given
	// sub-account
	ServiceAccounts int `json:"serviceAccounts,omitempty"`
	//Buckets is the number of buckets tied to the given sub-account
	Buckets int `json:"buckets,omitempty"`
	// Trial is the number of days remaining before trial expiration
	Trial int `json:"trial,omitempty"`
}

// Usage reports various bucket usage details and included fields will vary
// depending upon whether the query is for current usage or monthly usage.
type Usage struct {
	// Year only used in monthly usage report
	Year uint16 `json:"year,omitempty"`
	// Month only used in monthly usage report
	Month Month `json:"month,omitempty"`
	// NumBuckets only used in current usage report
	NumBuckets int `json:"numBuckets,omitempty"`
	// TotalUsageGB is the amount of space consumed in Gigabytes (hopefully)
	TotalUsageGB float64      `json:"totalUsageGB"`
	Buckets      []Bucket     `json:"buckets,omitempty"`
	SubAccounts  []SubAccount `json:"subAccounts,omitempty"`
}

// MonthlyUsageResp is the response object containing by month usage
// information by bucket. If the account performing the query is a sub-account,
// the SubAccounts field will not contain any data, because sub-accounts cannot
// contain sub-accounts.
type MonthlyUsageResp struct {
	// UsageByBucket contains a slice of Usage structs for all buckets under the master account or the sub-account for each month in the range.
	UsageByBucket []Usage `json:"usageByBucket,omitempty"`
	// UsageBySubAccount will be an empty slice unless the data is requested
	// with credentials belonging to a "master" account. In most instances data
	// will be queried with credentials belonging to a sub-account.
	UsageBySubAccount []Usage `json:"usageBySubAccount,omitempty"`
}

// CurrentUsageResp is the response object containing usage information by
// bucket. If the account performing the query is a sub-account, the SubAccounts
// field will not contain any data, because sub-accounts cannot contain
// sub-accounts.
type CurrentUsageResp struct {
	// UsageByBucket contains a slice of Usage structs for all buckets under the master account or the sub-account for each month in the range.
	UsageByBucket Usage `json:"usageByBucket,omitempty"`
	// UsageBySubAccount will be an empty slice unless the data is requested
	// with credentials belonging to a "master" account. In most instances data
	// will be queried with credentials belonging to a sub-account.
	UsageBySubAccount struct {
		TotalUsageGB float64      `json:"totalUsageGB"`
		SubAccounts  []SubAccount `json:"subAccounts,omitempty"`
	} `json:"usageBySubAccount,omitempty"`
}
