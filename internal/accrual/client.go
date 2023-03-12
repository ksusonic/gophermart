package accrual

import "net/http"

type Client struct {
	address string
	client  *http.Client
}

func NewClient(address string) *Client {
	return &Client{
		address: address,
		client:  &http.Client{},
	}
}

func (c *Client) getOrderInfo(number string) {
}
