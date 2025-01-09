package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	headerAccept        = "Accept"
	headerAuthorization = "Authorization"
	headerContent       = "Content-Type"
	mediaTypeJSON       = "application/json"
	mediaTypeURLForm    = "application/x-www-form-urlencoded"
	rateLimit           = 1 * time.Second
)

// Client is a Domain Management Backend API client
type Client struct {
	Endpoint string
	client   *http.Client
}

type rateLimitedTransport struct {
	delegate http.RoundTripper
	throttle time.Time
	sync.Mutex
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.Lock()
	defer t.Unlock()

	if t.throttle.After(time.Now()) {
		delta := time.Until(t.throttle)
		time.Sleep(delta)
	}

	t.throttle = time.Now().Add(rateLimit)
	return t.delegate.RoundTrip(req)
}

func NewClient(endpoint string) (*Client, error) {
	endpoint, err := formatURL(endpoint)
	if err != nil {
		return nil, err
	}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &Client{
		Endpoint: endpoint,
		client: &http.Client{
			Timeout: time.Second * 30,
			Transport: &rateLimitedTransport{
				delegate: netTransport,
				throttle: time.Now().Add(-(rateLimit)),
			},
		},
	}, nil
}

func formatURL(base string) (string, error) {
	endpoint, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	if endpoint.Host == "" || endpoint.Scheme == "" {
		return "", fmt.Errorf("invalid endpoint. expected format: scheme://host")
	}

	var finalPath string
	finalPath = fmt.Sprintf("%s://%s", endpoint.Scheme, endpoint.Host)

	// Some endpoints may have an extra path
	// Check for the existence of a path in
	if endpoint.Path != "" {
		finalPath = fmt.Sprintf("%s%s", finalPath, endpoint.Path)
	}

	return finalPath, err
}

func (c *Client) execute(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set(headerAccept, mediaTypeJSON)
	req.Header.Set(headerContent, mediaTypeURLForm)

	resp, err = c.client.Do(req)
	return
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
