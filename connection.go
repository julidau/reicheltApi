package reichelt

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type Connection struct {
	client http.Client

	queryCount int
}

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

	// get reichelt SID cookie set
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
