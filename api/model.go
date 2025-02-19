package api

import (
	"encoding/json"
	"net/url"
)

type DomainReq struct {
	Filter  Metadata
	Exclude Metadata
}

func (request *DomainReq) ToURLQuery() (url.Values, error) {
	filter := map[string]interface{}{
		"metadata": request.Filter,
	}
	exclude := map[string]interface{}{
		"metadata": request.Exclude,
	}

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	excludeBytes, err := json.Marshal(exclude)
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("filter", string(filterBytes))
	v.Set("exclude", string(excludeBytes))

	return v, nil
}

type Metadata struct {
	Labels      map[string]interface{} `json:"labels"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type Subdomain struct {
	Fqdn     string
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
