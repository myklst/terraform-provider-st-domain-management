package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// TODO Decide on getForTerraform or get only domain API. choose one
const (
	GetForTerraform         = "%s/domain/terraform"
	DomainAnnotations       = "%s/domain/%s/annotations"
	FullDomainsConfig       = "%s/domain/fulldomainsconfig"
	DomainAnnotationsDelete = "%s/domain/%s/annotations/delete"
)

func (c *Client) CreateAnnotations(domain string, payload []byte) (resp []byte, err error) {
	domainURL := fmt.Sprintf(DomainAnnotations, c.Endpoint, domain)

	req, err := http.NewRequest(http.MethodPost, domainURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.execute(req)
	if err != nil {
		return []byte(fmt.Sprintf("Client HTTP Error %s", err.Error())), err
	}

	defer httpResponse.Body.Close()
	if httpResponse.StatusCode >= 400 {
		body, _ := io.ReadAll(httpResponse.Body)
		return body, errors.New(strconv.Itoa(httpResponse.StatusCode))
	}
	return nil, nil
}

func (c *Client) ReadAnnotations(domain string, payload string) (resp []byte, err error) {
	domainURL := fmt.Sprintf(DomainAnnotations, c.Endpoint, domain)

	req, err := http.NewRequest(http.MethodGet, domainURL, bytes.NewBufferString(payload))
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.execute(req)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(httpResponse.Body)
}

func (c *Client) UpdateAnnotations(domain string, payload []byte) (resp []byte, err error) {
	domainURL := fmt.Sprintf(DomainAnnotations, c.Endpoint, domain)

	req, err := http.NewRequest(http.MethodPatch, domainURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.execute(req)
	if err != nil {
		responseError, _ := io.ReadAll(httpResponse.Body)
		return responseError, err
	}

	return nil, nil
}

func (c *Client) DeleteAnnotations(domain string, payload []byte) error {
	domainURL := fmt.Sprintf(DomainAnnotations, c.Endpoint, domain)

	req, err := http.NewRequest(http.MethodDelete, domainURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if _, err = c.execute(req); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFullDomainsConfig(domain string) (res *http.Response, err error) {
	domainURL := fmt.Sprintf(FullDomainsConfig, c.Endpoint)
	domainFilter := fmt.Sprintf(`{"domain":"%s"}`, domain)

	req, err := http.NewRequest(http.MethodGet, domainURL, strings.NewReader(domainFilter))
	if err != nil {
		return &http.Response{}, err
	}

	if res, err = c.execute(req); err != nil {
		return &http.Response{}, err
	}

	return
}

func (c *Client) GetForTerraform(payload bytes.Buffer) (res *http.Response, err error) {
	domainURL := fmt.Sprintf(GetForTerraform, c.Endpoint)

	req, err := http.NewRequest(http.MethodGet, domainURL, &payload)
	if err != nil {
		return &http.Response{}, err
	}

	if res, err = c.execute(req); err != nil {
		return &http.Response{}, err
	}

	return
}
