package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	GetDomainsFull = "%s/domains/full"
)

func (c *Client) GetDomainsFull(request DomainReq) (resp []byte, err error) {
	url, err := url.Parse(fmt.Sprintf(GetDomainsFull, c.Endpoint))
	if err != nil {
		return nil, err
	}

	query, err := request.ToURLQuery()
	if err != nil {
		return nil, err
	}
	url.RawQuery = query.Encode()

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
		// so that TF can proceed with warning.
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

	commonResp := commonResponse{}
	if err = json.Unmarshal(body, &commonResp); err != nil {
		return nil, err
	}

	return commonResp.Dt, nil
}
