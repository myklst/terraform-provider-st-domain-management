package api

import (
	"encoding/json"
	"net/url"
)

type DomainReq struct {
	FilterDomains    *IncludeExclude `json:"domains,omitempty"`
	FilterSubdomains *IncludeExclude `json:"subdomains,omitempty"`
}

func (request *DomainReq) ToURLQuery() (url.Values, error) {
	v := url.Values{}

	if request == nil {
		return v, nil
	}

	filter, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	if len(filter) > 0 {
		v.Set("filter", string(filter))
	}

	return v, nil
}

type IncludeExclude struct {
	Include *Include `json:"include,omitempty"`
	Exclude *Exclude `json:"exclude,omitempty"`
}

type Include struct {
	Metadata *Metadata `json:"metadata" binding:"omitempty"`
}

type Exclude struct {
	Metadata *Metadata `json:"metadata" binding:"omitempty"`
}

type Metadata struct {
	Labels      map[string]interface{} `json:"labels,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type Subdomain struct {
	Fqdn     string   `json:"fqdn,omitempty"`
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

type Domain struct {
	Domain   string   `json:"domain"`
	Metadata Metadata `json:"metadata"`
}

type DomainFull struct {
	Domain     string      `json:"domain"`
	Metadata   Metadata    `json:"metadata"`
	Subdomains []Subdomain `json:"subdomains"`
}

type DomainResponse struct {
	Domains []*Domain `json:"dt"`
}

type DomainFullResponse struct {
	DomainsFull []*DomainFull `json:"dt"`
}

type AnnotationsResponse struct {
	Domain Domain `json:"dt"`
}
