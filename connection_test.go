package reichelt

import (
	"image/jpeg"
	"log"
	"net/url"
	"testing"
)

var conn *Connection

func SetupConnection(t *testing.T) {
	c, err := NewConnection()
	if err != nil {
		t.Fatal("create Connection:", err)
	}

	// look for cookie
	url, _ := url.Parse(apiurl)
	cookies := c.Jar.Cookies(url)
	found := false
	for _, k := range cookies {
		t.Log("cookie set:", k)
		if k.Name == "Reichelt_SID" {
			found = true
			break
		}
	}

	if !found {
		log.Fatal("connection did not get cookie")
	}

	t.Log("connection created successfully")
	conn = c
}

func TestPart(t *testing.T) {
	t.Run("createConnection", SetupConnection)

	// get part
	parts, err := conn.FindPart("1N4001")
	if err != nil {
		t.Fatal("find part:", err)
	}

	if len(parts) == 0 {
		log.Fatal("not enough parts were retrieved")
	}

	t.Log(parts)

	// get prices
	p := conn.GetPrice(parts[0])
	if p == 0 {
		t.Fatal("get Price")
	}

	t.Log(parts[0], ":", p)

	// get image for part
	imgReader, err := conn.GetImage(parts[0], 1000, 1000)
	if err != nil {
		t.Fatal("get product image:", err)
	}
	defer imgReader.Close()
	img, err := jpeg.Decode(imgReader)

	if err != nil {
		t.Fatal("jpg decode:", err)
	}

	t.Log("image size:", img.Bounds())
}
