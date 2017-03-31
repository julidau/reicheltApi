package reichelt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"

	"golang.org/x/net/html"
)

type Part struct {
	Number      int
	Description string
}

// used for the json Decoder
// The fields are then copied to Part
// (This was done to remove the structure tags).
type partInternal struct {
	Number      int    `json:"article_artid"`
	Description string `json:"article_lang_besch"`
}

// Where to dispatch api queries to
const apiurl = "https://www.reichelt.de/index.html"

// One Response-field from the autocomplete
// endpoint (a small excerpt from it)
type responseField struct {
	NumFound int     `json:"numFound"`
	MaxScore float32 `json:"maxScore"`

	Docs []partInternal `json:"docs"`
}

// The Full response
// notice: There are more fields that the server response contains
type searchResponse struct {
	Response responseField `json:"response"`
}

var priceSelector = cascadia.MustCompile("#av_price")

// Search for a part like using the sites search engine
// can be used to resolv partnumbers to internal ones
func (c *Connection) FindPart(query string) (result []Part, err error) {
	resp, err := c.client.Get(apiurl + "?ACTION=514&id=8&term=" + url.PathEscape(query))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong response status: %d", resp.StatusCode)
	}

	reader := json.NewDecoder(resp.Body)
	response := searchResponse{}
	if err = reader.Decode(&response); err != nil {
		return nil, err
	}

	result = make([]Part, len(response.Response.Docs))

	for i, j := range response.Response.Docs {
		result[i] = Part{
			Number:      j.Number,
			Description: j.Description,
		}
	}

	return result, nil
}

// Returns the Price of the Part
// or 0 if there was an error
func (c *Connection) GetPrice(p Part) float32 {
	resp, err := c.client.Get(apiurl + "?ACTION=3&ARTICLE=" + strconv.Itoa(p.Number))

	if err != nil {
		//		log.Println("price:", "get request:", err)
		return 0
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//		log.Println("price:", "wrong result:", resp.Status)
		return 0
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		//		log.Println("price:", "parse html:", err)
		return 0
	}
	priceTag := priceSelector.MatchFirst(doc)

	if priceTag == nil {
		//		log.Println("price:", "selector returned nothing")
		return 0
	}

	// retrieve first child of node
	// since inner Text is saved as child node
	// NOTE: This might be the wrong node, but if it is,
	// code below this one will fail anyway, so we dont check
	// the node type here
	price := priceTag.FirstChild.Data

	// split before € sign
	i := strings.Index(price, " €")
	if i == len(price) || len(price) == 0 {
		return 0
	}

	// need to convert german float (using ,) to american decimals
	// using . for ParseFloat (since ParseFloat is not localized)
	str := strings.Replace(price[:i], ",", ".", 1)
	ret, _ := strconv.ParseFloat(str, 32)
	return float32(ret)
}
