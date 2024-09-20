package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	GetOnlyDomain     = "%s/domains/fqdn"
	GetDomainsFull    = "%s/domains/full"
	DomainAnnotations = "%s/domains/%s/annotations"
)

func (c *Client) CreateAnnotations(domain string, payload []byte) (resp []byte, err error) {
	url, err := url.Parse(fmt.Sprintf(DomainAnnotations, c.Endpoint, domain))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(payload))
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

func (c *Client) ReadAnnotations(domain string, payload []byte) (resp []byte, err error) {
	url, err := url.Parse(fmt.Sprintf(DomainAnnotations, c.Endpoint, domain))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("filter", string(payload))
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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
	url, err := url.Parse(fmt.Sprintf(DomainAnnotations, c.Endpoint, domain))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.execute(req)
	if err != nil {
		return []byte(fmt.Sprintf("Client HTTP Error %s", err.Error())), err
	}

	defer httpResponse.Body.Close()
	body, _ := io.ReadAll(httpResponse.Body)

	if httpResponse.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(httpResponse.StatusCode))
	}

	return body, nil
}

func (c *Client) DeleteAnnotations(domain string, payload []byte) (resp []byte, err error) {
	url, err := url.Parse(fmt.Sprintf(DomainAnnotations, c.Endpoint, domain))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("filter", string(payload))
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.execute(req)
	if err != nil {
		return []byte(fmt.Sprintf("Client HTTP Error %s", err.Error())), err
	}

	defer httpResponse.Body.Close()
	body, _ := io.ReadAll(httpResponse.Body)

	if httpResponse.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(httpResponse.StatusCode))
	}
	return body, nil
}

func (c *Client) GetOnlyDomain(payload []byte) (res *http.Response, err error) {
	url, err := url.Parse(fmt.Sprintf(GetOnlyDomain, c.Endpoint))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("filter", string(payload))
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	if res, err = c.execute(req); err != nil {
		return &http.Response{}, err
	}

	return
}

func (c *Client) GetDomainsFull(payload []byte) (res *http.Response, err error) {
	url, err := url.Parse(fmt.Sprintf(GetDomainsFull, c.Endpoint))
	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("filter", string(payload))
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	if res, err = c.execute(req); err != nil {
		return &http.Response{}, err
	}

	return
}
