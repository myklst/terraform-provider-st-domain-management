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

func (c *Client) CreateAnnotations(domain string, payload string) (resp []byte, err error) {
	path, err := url.JoinPath(c.Endpoint, "domains", domain, "annotations")
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(path)
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
	path, err := url.JoinPath(c.Endpoint, "domains", domain, "annotations")
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(path)
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

	if metadata.Domain.Metadata.Annotations != nil {
		return nil, nil
	}

	return metadata.Domain.Metadata.Annotations, nil
}

func (c *Client) UpdateAnnotations(domain string, payload []byte) (resp []byte, err error) {
	path, err := url.JoinPath(c.Endpoint, "domains", domain, "annotations")
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(path)
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
	path, err := url.JoinPath(c.Endpoint, "domains", domain, "annotations")
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(path)
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
