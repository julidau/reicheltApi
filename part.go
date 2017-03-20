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
	Number int `json:"article_artid"`

	Description string `json:"article_lang_besch"`
}

const apiurl = "https://www.reichelt.de/index.html"

type ResponseField struct {
	NumFound int     `json:"numFound"`
	MaxScore float32 `json:"maxScore"`

	Docs []Part `json:"docs"`
}

type SearchResponse struct {
	Response ResponseField `json:"response"`
}

var (
	priceSelector = cascadia.MustCompile("#av_price")
)

// Search for a part like using the sites search engine
// can be used to resolv partnumbers to internal ones
func (c *Connection) FindPart(query string) ([]Part, error) {
	resp, err := c.client.Get(apiurl + "?ACTION=514&id=8&term=" + url.PathEscape(query))
	c.queryCount++

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong response status: %d", resp.StatusCode)
	}

	reader := json.NewDecoder(resp.Body)
	response := SearchResponse{}
	if err = reader.Decode(&response); err != nil {
		return nil, err
	}

	return response.Response.Docs, nil
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

	// need to convert german decimals (using ,) to american decimals
	// using .
	str := strings.Replace(price[:i], ",", ".", 1)
	ret, _ := strconv.ParseFloat(str, 32)
	return float32(ret)
}
