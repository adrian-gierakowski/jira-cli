package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TenantInfo holds tenant information.
type TenantInfo struct {
	CloudID string `json:"cloudId"`
}

// TenantInfo gets tenant information.
func (c *Client) TenantInfo() (*TenantInfo, error) {
	res, err := c.request(context.Background(), http.MethodGet, fmt.Sprintf("%s/_edge/tenant_info", c.server), nil, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var tenantInfo TenantInfo
	if err := json.NewDecoder(res.Body).Decode(&tenantInfo); err != nil {
		return nil, err
	}

	return &tenantInfo, nil
}
