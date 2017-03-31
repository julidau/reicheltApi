package reichelt

import (
	"io"
	"strconv"
)

// gets the product image of a reichelt article using the internal Part number
// the reader will return a image/jpg file
// NOTE: the Reader must be closed
func (c *Connection) GetImage(p Part, w, h uint) (io.ReadCloser, error) {
	resp, err := c.client.Get("https://www.reichelt.de/artimage/resize_" + strconv.Itoa(int(w)) + "x" + strconv.Itoa(int(h)) + "/" + strconv.Itoa(p.Number))

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
