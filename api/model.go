package api

import (
	"encoding/json"
	"net/url"
)

type DomainReq struct {
	FilterDomains    *IncludeExclude `json:"filter_domains,omitempty"`
	FilterSubdomains *IncludeExclude `json:"filter_subdomains,omitempty"`
}

func (request *DomainReq) ToURLQuery() (url.Values, error) {
	v := url.Values{}

	if request.FilterDomains != nil {
		filterDomains, err := json.Marshal(request.FilterDomains)
		if err != nil {
			return nil, err
		}

		if len(filterDomains) > 2 {
			v.Set("filter_domains", string(filterDomains))
		}
	}

	if request.FilterSubdomains != nil {
		filterSubdomains, err := json.Marshal(request.FilterSubdomains)
		if err != nil {
			return nil, err
		}

		if len(filterSubdomains) > 2 {
			v.Set("filter_domains", string(filterSubdomains))
		}
	}

	return v, nil
}

type IncludeExclude struct {
	Include *Metadata `json:"include,omitempty"`
	Exclude *Metadata `json:"exclude,omitempty"`
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
	DomainsFull []*DomainFull
}

type AnnotationsResponse struct {
	Domain Domain `json:"dt"`
}

type commonResponse struct {
	Dt  json.RawMessage
	Err json.RawMessage
}
