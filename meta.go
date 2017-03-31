package reichelt

import (
	"strconv"
	"strings"

	"github.com/andybalholm/cascadia"

	"golang.org/x/net/html"
)

type Meta map[string]map[string]string

var (
	metaSelector          = cascadia.MustCompile(".av_propview")
	metaItemNameSelector  = cascadia.MustCompile(".av_propname")
	metaItemValueSelector = cascadia.MustCompile(".av_propvalue")
	mpnSelector           = cascadia.MustCompile("#av_articlemanufacturer > .av_fontnormal")
	manufacturerSelector  = cascadia.MustCompile("#av_articlemanufacturer > [itemprop=\"manufacturer\"]")

	datasheetSelector = cascadia.MustCompile(".av_datasheet_description a")
)

// Get Metadata connected to specified part
func (c *Connection) GetMeta(p Part) (Meta, error) {
	resp, err := c.client.Get("https://www.reichelt.de/index.html?ACTION=3&ARTICLE=" + strconv.Itoa(p.Number))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)

	if err != nil {
		return nil, err
	}

	nodes := metaSelector.MatchAll(doc)
	if nodes == nil {
		return nil, nil
	}

	result := make(Meta)

	for _, n := range nodes {
		if n.FirstChild == nil || n.FirstChild.FirstChild == nil {
			continue
		}

		headline := n.FirstChild.FirstChild.Data
		data := make(map[string]string)

		names := metaItemNameSelector.MatchAll(n)
		values := metaItemValueSelector.MatchAll(n)

		if len(names) != len(values) {
			continue
		}

		for i := range names {
			if names[i].FirstChild == nil || values[i].FirstChild == nil {
				continue
			}

			data[names[i].FirstChild.Data] = strings.Trim(values[i].FirstChild.Data, " ")
		}

		result[headline] = data
	}

	// insert datasheets
	nodes = datasheetSelector.MatchAll(doc)
	temp := make(map[string]string)

	for _, node := range nodes {

		if node.FirstChild == nil || node.FirstChild.Type != html.TextNode {
			continue
		}

		name := node.FirstChild.Data
		link := ""

		// find href of link
		for _, k := range node.Attr {
			if k.Key == "href" {
				link = k.Val
			}
		}

		if link == "" {
			// no link found
			continue
		}

		temp[name] = link
	}

	result["datasheets"] = temp

	// get MPN
	node := mpnSelector.MatchFirst(doc)
	if node != nil && node.FirstChild != nil {
		temp["mpn"] = node.FirstChild.Data
	}

	// get Manufacturer
	node = manufacturerSelector.MatchFirst(doc)
	if node != nil && node.FirstChild != nil {
		temp["manufacturer"] = node.FirstChild.Data
	}

	return result, nil
}
