package api

type Metadata struct {
	Labels      map[string]interface{} `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
}

type Subdomain struct {
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

type Domain struct {
	Domain   string   `json:"domain"`
	Metadata Metadata `json:"metadata"`
}

type DomainFull struct {
	Domain     string       `json:"domain"`
	Metadata   Metadata     `json:"metadata"`
	Subdomains []*Subdomain `json:"subdomains"`
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
