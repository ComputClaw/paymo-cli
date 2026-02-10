package api

// GetClients returns all clients
func (c *Client) GetClients() ([]PaymoClient, error) {
	var resp ClientsResponse
	if err := c.Get("clients", &resp); err != nil {
		return nil, err
	}

	return resp.Clients, nil
}
