package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"

	"go.dedaa.de/julixau/reichelt"
)

var (
	addr = flag.String("http", ":8080", "The Address to bind to ")
)

type Handler struct {
	*reichelt.Connection

	cache *cache.Cache
}

func NotFound(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusNotFound)
	fmt.Fprint(resp, "404")
}

func InternalError(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(resp, "500")
}

func (h Handler) Search(resp http.ResponseWriter, path []string) {
	if len(path) == 0 {
		NotFound(resp)
		return
	}
	log.Println("level 2 request:", path)

	query, err := url.PathUnescape(path[0])
	if err != nil {
		log.Println("illegal query:", path[0], ":", err)
		NotFound(resp)
		return
	}

	parts, err := h.FindPart(query)
	if err != nil {
		log.Println("error retrieving part:", err)
		InternalError(resp)
	}

	// insert to cache
	if h.cache != nil {
		for _, p := range parts {
			h.cache.Set(strconv.Itoa(p.Number), &p, cache.NoExpiration)
		}
	}

	encoder := json.NewEncoder(resp)
	encoder.Encode(parts)
}

func (h Handler) Picture(resp http.ResponseWriter, path []string) {
	if len(path) == 0 {
		NotFound(resp)
		return
	}
	log.Println("level 2 request:", path)
	serve := func(img image.Image) {
		resp.Header().Set("Content-type", "image/png")
		resp.WriteHeader(http.StatusOK)
		if err := png.Encode(resp, img); err != nil {
			log.Println("could not encode png:", err)
		}
	}

	number, err := strconv.Atoi(path[0])
	if err != nil {
		log.Println("encountered decode error:", err)
		NotFound(resp)
	}

	if h.cache != nil {
		if x, ok := h.cache.Get(path[0] + "-image"); ok {
			serve(*(x.(*image.Image)))
			return
		}
	}

	img, err := h.GetImage(reichelt.Part{Number: number}, 99999, 9999)

	if err != nil {
		log.Println("error retrieving picture:", err)
		InternalError(resp)
		return
	}

	decodedImg, err := jpeg.Decode(img)
	if err != nil {
		log.Println("could no decode image:", err)
		InternalError(resp)
		return
	}
	if h.cache != nil {
		h.cache.Set(path[0]+"-image", &decodedImg, cache.NoExpiration)
	}
	serve(decodedImg)
}

func (h Handler) Price(resp http.ResponseWriter, path []string) {
	if len(path) == 0 {
		NotFound(resp)
		return
	}

	log.Println("level 2 request:", path)

	number, err := strconv.Atoi(path[0])
	if err != nil {
		log.Println("encountered decode error:", err)
		NotFound(resp)
	}
	var price float32

	if h.cache != nil {
		if x, ok := h.cache.Get(path[0] + "-price"); ok {
			price = x.(float32)
			goto cached
		}
	}

	price = h.GetPrice(reichelt.Part{Number: number})
	if h.cache != nil {
		h.cache.Set(path[0]+"-price", price, time.Second*30)
	}

cached:
	encoder := json.NewEncoder(resp)
	encoder.Encode(price)
}

func (h Handler) Meta(resp http.ResponseWriter, path []string) {
	if len(path) == 0 {
		NotFound(resp)
		return
	}

	log.Println("level 2 request:", path)

	number, err := strconv.Atoi(path[0])
	if err != nil {
		log.Println("encountered decode error:", err)
		NotFound(resp)
	}

	// implement caching to avoid many queries to reichelt server
	var meta reichelt.Meta
	if h.cache != nil {
		if x, ok := h.cache.Get(path[0] + "-meta"); ok {
			meta = x.(reichelt.Meta)
			goto cached
		}
	}

	meta, err = h.GetMeta(reichelt.Part{Number: number})
	if err != nil {
		log.Println("encountered error:", err)
		InternalError(resp)
	}
	if h.cache != nil {
		h.cache.Set(path[0]+"-meta", meta, cache.NoExpiration)
	}

cached:
	encoder := json.NewEncoder(resp)

	if len(path) > 1 {
		if strings.ToLower(path[1]) == "overview" {
			var headlines []string
			for k, _ := range meta {
				headlines = append(headlines, k)
			}
			encoder.Encode(headlines)
			return
		}

		if query, err := url.QueryUnescape(path[1]); err != nil {
			NotFound(resp)
			log.Println("illegal query:", path[1])
		} else {
			// query is more concrete
			subset, ok := meta[query]
			if !ok {
				NotFound(resp)
				return
			}
			encoder.Encode(subset)
		}
	} else {
		encoder.Encode(meta)
	}

}

func (h Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// find out whether there was URL encoded data in the query
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	p := strings.Split(path, "/")
	if len(p) < 2 {
		NotFound(resp)
		return
	}
	p = p[1:]
	log.Println("level 1 request:", p)

	switch p[0] {
	case "search":
		h.Search(resp, p[1:])
	case "image":
		h.Picture(resp, p[1:])
	case "price":
		h.Price(resp, p[1:])
	case "meta":
		h.Meta(resp, p[1:])
	default:
		NotFound(resp)
	}
}

// a Simple request server
// exposing a simple api
// to search and retrieve
//  - price
//  - productimage
// for a product
func main() {
	flag.Parse()
	conn, err := reichelt.NewConnection()
	if err != nil {
		log.Fatal("could not create connection to reichelt:", err)
	}

	log.Println("start serving on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, Handler{conn, cache.New(cache.NoExpiration, 0)}))
}
