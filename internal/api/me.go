package api

// GetMe returns the current authenticated user
func (c *Client) GetMe() (*User, error) {
	var resp MeResponse
	if err := c.Get("me", &resp); err != nil {
		return nil, err
	}
	
	if len(resp.Users) == 0 {
		return nil, &APIError{StatusCode: 404, Message: "no user found"}
	}
	
	return &resp.Users[0], nil
}

// ValidateAuth checks if the current authentication is valid
func (c *Client) ValidateAuth() error {
	_, err := c.GetMe()
	return err
}