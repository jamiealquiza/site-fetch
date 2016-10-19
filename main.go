package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// Runtime configs, set by flags.
var config struct {
	url   string
	depth int
}

// SiteMap is a site map.
var SiteMap = struct {
	sync.Mutex
	Map map[string]PageMap
}{Map: make(map[string]PageMap)}

// Crawl request options. Depth
// is used as a decrementing counter,
// stripped off at each level of recursion
// with a stop-at-zero.
type CrawlRequest struct {
	Url   string
	Depth int
}

// A PageMap{} includes a list of links and
// static assets for a given URL.
type PageMap struct {
	url    string
	Assets []string `json:"assets"`
	Links  []string `json:"links"`
}

func init() {
	flag.StringVar(&config.url, "url", "", "Target URL")
	flag.IntVar(&config.depth, "depth", 1, "Recursive crawl depth")
	flag.Parse()

	if config.url == "" {
		fmt.Println("Please enter a target URL")
		os.Exit(1)
	}

	// Update config.url if the originally requested
	// URL results in a redirect.
	resp, err := http.Head(config.url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	config.url = strings.Trim(resp.Request.URL.String(), "/")

}

func main() {
        siteMap := crawl(&CrawlRequest{Url: config.url, Depth: config.depth})

        siteMapJson, _ := json.MarshalIndent(siteMap, "", "  ")
        fmt.Printf("%s\n", siteMapJson)
}

// crawl takes a URL, fetches the page, parses the page HTML,
// tokenizes the HTML with tokenize(), then passes the tokens
// to mapAssetsAndLinks(). This returns a PageMap{} that is
// then populated in the final SiteMap.Map map with the URL
// as the map key.
func crawl(request *CrawlRequest) map[string]PageMap {
	// Break if we've recursed to the max depth.
	if request.Depth <= 0 {
		return SiteMap.Map
	}
	request.Depth--

	// Break if we've already crawled this URL.
	if _, exists := SiteMap.Map[request.Url]; exists {
		return SiteMap.Map
	}

	// Fetch page.
	resp, err := http.Get(request.Url)
	if err != nil {
		fmt.Println(err)
		if request.Depth+1 == config.depth {
			os.Exit(1)
		}
		return SiteMap.Map
	}
	defer resp.Body.Close()

	// Skip non 200.
	if resp.Status != "200 OK" {
		return SiteMap.Map
	}

	// Write page resp to a buffer since the html 
	//package tokenizer requires an io.Reader.
	var respBuf bytes.Buffer
	io.Copy(&respBuf, resp.Body)

	// Tokenize page.
	tokens := tokenize(&respBuf)

	// Use tokens to get a page map of links and assets.
	pageMap := mapAssetsAndLinks(tokens, config.url)

	// Add the page map of links and assets
	// to SiteMap.Map with the reference URL as the key.
	SiteMap.Lock()
	SiteMap.Map[request.Url] = pageMap
	SiteMap.Unlock()

	wg := &sync.WaitGroup{}

	// Now recursely crawl each link in pageMap.Links to the defined request.Depth.
	for _, url := range pageMap.Links {
		urlCrawlRequest := &CrawlRequest{Url: url, Depth: request.Depth}

		wg.Add(1)
		go asyncCrawl(urlCrawlRequest, wg)
	}

	wg.Wait()
	return SiteMap.Map
}

func asyncCrawl(request *CrawlRequest, wg *sync.WaitGroup) {
	crawl(request)
	wg.Done()
}

// tokenize takes a page request buffer and returns
// a slice of parsed HTML tokens.
func tokenize(page io.Reader) []html.Attribute {
	t := html.NewTokenizer(page)
	tokens := make([]html.Attribute, 0)

	// Iteratively gets tokens.
	for {
		tokenType := t.Next()
		token := t.Token()
		if tokenType == html.ErrorToken {
			// We're done.
			break
		}

		// Tokens come back in a slice.
		// For simplicity, we populate a single slice (tokens)
		// for return - rather than a multi-dimensional slice.
		for _, i := range token.Attr {
			tokens = append(tokens, i)
		}
	}

	return tokens
}

// mapAssetsAndLinks takes a slice of tokens from tokenize()
// and returns a PageMap{}. A PageMap{} includes a list of
// links and static assets for a given URL.
func mapAssetsAndLinks(t []html.Attribute, url string) PageMap {
	pm := PageMap{}
	pm.url = url

	// We use a map to build a set for each object
	// type we're interested in (links, assets).
	// For instance, if we discover the link 'somedomain.com/about',
	// we assign it to objects["somedomain.com/about"] = "link".
	// This ensures each 'object' is being collected once, at most.
	objects := make(map[string]string)

	// Assign found assets and links to map
	// to remove duplicates.
	for _, token := range t {
		// Remove query strings and trailing slashes
		// for any links or assets found.
		switch token.Key {

		// Links.
		case "href":
			// Don't include references to home.
			if token.Val == "/" {
				continue
			}
			// Only include references to this domain.
			if strings.HasPrefix(token.Val, "/") {
				object := sanitize(token.Val)
				objects[url+"/"+object] = "link"
			} else if strings.HasPrefix(token.Val, url) {
				object := sanitize(token.Val)
				objects[object] = "link"
			}

		// Assets.
		case "src":
			if strings.HasPrefix(token.Val, "/") {
				object := sanitize(token.Val)
				objects[url+"/"+object] = "asset"
			} else {
				object := sanitize(token.Val)
				objects[object] = "asset"
			}
		}
	}

	// Once all objects have been found, 
	// iterate object map and populate into a PageMap.
	for value, t := range objects {
		switch t {
		case "link":
			pm.Links = append(pm.Links, value)
		case "asset":
			pm.Assets = append(pm.Assets, value)
		}
	}

	return pm
}

func sanitize(u string) string {
	sanitized := strings.Split(u, "?")[0]
	sanitized = strings.Trim(sanitized, "/")
	return sanitized
}
