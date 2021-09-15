// Code generated by goa v3.5.2, DO NOT EDIT.
//
// reporter client
//
// Command:
// $ goa gen github.com/lbryio/lbrytv/apps/watchman/design -o apps/watchman

package reporter

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Client is the "reporter" service client.
type Client struct {
	AddEndpoint     goa.Endpoint
	HealthzEndpoint goa.Endpoint
}

// NewClient initializes a "reporter" service client given the endpoints.
func NewClient(add, healthz goa.Endpoint) *Client {
	return &Client{
		AddEndpoint:     add,
		HealthzEndpoint: healthz,
	}
}

// Add calls the "add" endpoint of the "reporter" service.
func (c *Client) Add(ctx context.Context, p *PlaybackReport) (err error) {
	_, err = c.AddEndpoint(ctx, p)
	return
}

// Healthz calls the "healthz" endpoint of the "reporter" service.
func (c *Client) Healthz(ctx context.Context) (res string, err error) {
	var ires interface{}
	ires, err = c.HealthzEndpoint(ctx, nil)
	if err != nil {
		return
	}
	return ires.(string), nil
}
