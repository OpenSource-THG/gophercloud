package limits

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// A limit is the limit that override the registered limit for each project.
type Limit struct {
	// ID is the unique ID of the limit.
	ID string `json:"id"`

	// RegionID is the ID of the region where the limit is applied.
	RegionID string `json:"region_id"`

	// ProjectID is the ID of the project where the limit is applied.
	ProjectID string `json:"project_id"`

	// DomainID is the ID of the domain where the limit is applied.
	DomainID string `json:"domain_id"`

	// ServiceID is the ID of the service where the limit is applied.
	ServiceID string `json:"service_id"`

	// Description of the limit.
	Description string `json:"description"`

	// ResourceName is the name of the resource that the limit is applied to.
	ResourceName string `json:"resource_name"`

	// ResourceLimit is the override limit.
	ResourceLimit int `json:"resource_limit"`

	// Links contains referencing links to the limit.
	Links map[string]interface{} `json:"links"`
}

type LimitOutput struct {
	Limit *Limit `json:"limit"`
}

type LimitsOutput struct {
	Limits []Limit `json:"limits"`
}

type commonResult struct {
	gophercloud.Result
}

// CreateResult is the response from a Create operation. Call its Extract method
// to interpret it as a Limits.
type CreateResult struct {
	gophercloud.Result
}

// GetResult is the response from a Get operation. Call its Extract method
// to interpret it as a Limit.
type GetResult struct {
	commonResult
}

// UpdateResult is the result of an Update request. Call its Extract method to
// interpret it as a Limit.
type UpdateResult struct {
	commonResult
}

// DeleteResult is the response from a Delete operation. Call its ExtractErr to
// determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// LimitPage is a single page of Limit results.
type LimitPage struct {
	pagination.LinkedPageBase
}

// IsEmpty determines whether or not a page of Limits contains any results.
func (r LimitPage) IsEmpty() (bool, error) {
	limits, err := ExtractLimits(r)
	return len(limits) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (r LimitPage) NextPageURL() (string, error) {
	var s struct {
		Links struct {
			Next     string `json:"next"`
			Previous string `json:"previous"`
		} `json:"links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return s.Links.Next, err
}

// ExtractLimits returns a slice of Limits contained in a single page of
// results.
func ExtractLimits(r pagination.Page) ([]Limit, error) {
	var out LimitsOutput
	err := (r.(LimitPage)).ExtractInto(&out)
	return out.Limits, err
}

// Extract interprets any commonResult as a Limit.
func (r commonResult) Extract() (*Limit, error) {
	var out LimitOutput
	err := r.ExtractInto(&out)
	return out.Limit, err
}

// Extract interprets CreateResult as slice of Limits.
func (r CreateResult) Extract() ([]Limit, error) {
	var out LimitsOutput
	err := r.ExtractInto(&out)
	return out.Limits, err
}
