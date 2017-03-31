package reichelt

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type Connection struct {
	client http.Client
}

// Opens a new connection to the reichelt-Server
// this will try to connect and consequently
// throw an error on connection failure
func NewConnection() (c *Connection, err error) {
	jar, err := cookiejar.New(nil)

	if err != nil {
		return nil, err
	}

	c = &Connection{
		client: http.Client{
			Jar: jar,
		},
	}

	// set reichelt SID cookie
	resp, err := c.client.Get(apiurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Wrong Status response: %d(%s)", resp.StatusCode, resp.Status)
	}

	return c, nil
}
