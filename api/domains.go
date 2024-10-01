package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	GetDomain         = "%s/domains"
	GetDomainsFull    = "%s/domains/full"
	DomainAnnotations = "%s/domains/%s/annotations"
)

func (c *Client) CreateAnnotations(domain string, payload string) (resp []byte, err error) {
	url, err := url.Parse(fmt.Sprintf(DomainAnnotations, c.Endpoint, domain))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer([]byte(payload)))
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

func (c *Client) ReadAnnotations(domain string, payload []byte) (resp map[string]any, err error) {
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

	var httpResp *http.Response
	if httpResp, err = c.execute(req); err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		// If no annotations are found, dont return error,
		// so that TF can proceed with plan with empty annotations as input.
		if httpResp.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		return nil, handleErrorResponse(httpResp)
	}

	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var metadata AnnotationsResponse
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}

	if len(metadata.Domain.Metadata.Annotations) == 0 {
		return nil, nil
	}

	return metadata.Domain.Metadata.Annotations, nil
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

func (c *Client) GetDomains(payload []byte) (resp []*Domain, err error) {
	url, err := url.Parse(fmt.Sprintf(GetDomain, c.Endpoint))
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

	var httpResp *http.Response
	if httpResp, err = c.execute(req); err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		// If no domains are found, dont return error,
		// so that TF can proceed with plan with empty domains as input.
		if httpResp.StatusCode == http.StatusBadRequest {
			return nil, nil
		}

		return nil, handleErrorResponse(httpResp)
	}

	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	respData := DomainResponse{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, err
	}

	return respData.Domains, nil
}

func (c *Client) GetDomainsFull(payload []byte) (resp []*DomainFull, err error) {
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

	var httpResp *http.Response
	if httpResp, err = c.execute(req); err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		// If no domains are found, dont return error,
		// so that TF can proceed with plan with empty domains as input.
		if httpResp.StatusCode == http.StatusBadRequest {
			return nil, nil
		}

		return nil, handleErrorResponse(httpResp)
	}

	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	respData := DomainFullResponse{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, err
	}

	return respData.DomainsFull, nil
}

func handleErrorResponse(resp *http.Response) error {
	var body []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	jsonBody := map[string]interface{}{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		return err
	}

	return fmt.Errorf(
		"Got response %s: %s",
		strconv.Itoa(resp.StatusCode),
		jsonBody["err"],
	)
}
