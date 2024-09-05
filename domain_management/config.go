package domain_management

import (
	"fmt"
	"github.com/myklst/terraform-provider-st-domain-management/api"
)

type Config struct {
	Endpoint string
}

func (c *Config) Client() (*api.Client, error) {
	client, err := api.NewClient(c.Endpoint)

	if err != nil {
		return nil, fmt.Errorf("error setting up client: %s", err)
	}

	return client, nil
}
