package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

var links = make([]string, 0)
var visited = make(map[string]bool)

// CrawlWebpage crawls the given rootURL looking for <a href=""> tags
// that are targeting the current web page, either via an absolute url like http://mysite.com/mypath or by a relative url like /mypath
// and returns a sorted list of absolute urls  (e.g., []string{"http://mysite.com/1","http://mysite.com/2"})
func CrawlWebpage(rootURL string, maxDepth int) ([]string, error) {
	if rootURL == "" {
		return nil, fmt.Errorf("rootURL cannot be empty")
	}
	log.Println("Started Crawling...")
	crawl(rootURL, rootURL, maxDepth)
	sort.Strings(links)
	return links, nil
}

func crawl(rootURL string, baseURL string, depth int) {
	if depth < 0 {
		return
	}
	resp, err := http.Get(baseURL)
	if err != nil {
		log.Println("Error fetching", baseURL, ":", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Non-OK status code:", resp.Status)
		return
	}

	tokenizer := html.NewTokenizer(resp.Body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						link := attr.Val
						absoluteURL, err := makeAbsoluteURL(link, rootURL)
						if err == nil {
							if sameHost(absoluteURL, rootURL) {
								if !visited[absoluteURL] {
									visited[absoluteURL] = true
									links = append(links, absoluteURL)
									if (depth - 1) > 0 {
										crawl(rootURL, absoluteURL, depth-1)
									}
								}
							}
						} else {
							log.Println("Error creating absolute URL:", err)
						}
					}
				}
			}
		}
	}
}

func sameHost(url1 string, url2 string) bool {
	parsedURL1, err := url.Parse(url1)
	if err != nil {
		log.Println("Error parsing URL1:", err)
		return false
	}

	parsedURL2, err := url.Parse(url2)
	if err != nil {
		log.Println("Error parsing URL2:", err)
		return false
	}
	// Extracting main domain from the host
	domain1 := getMainDomain(parsedURL1.Host)
	domain2 := getMainDomain(parsedURL2.Host)

	// Comparing the main domains
	return domain1 == domain2
}

func getMainDomain(host string) string {
	// Remove "www." if present and split by "."
	parts := strings.Split(strings.TrimPrefix(host, "www."), ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

func makeAbsoluteURL(relativeURL string, incomingURL string) (string, error) {
	// Return if anything other than http or https
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL, nil
	}

	base, err := url.Parse(incomingURL)
	if err != nil {
		return "", err
	}

	absoluteURL := base.Scheme + "://" + base.Host + relativeURL
	return path.Clean(absoluteURL), nil
}

// --- DO NOT MODIFY BELOW ---

func main() {
	const (
		defaultURL      = "https://www.example.com/"
		defaultMaxDepth = 3
	)
	urlFlag := flag.String("url", defaultURL, "the url that you want to crawl")
	maxDepth := flag.Int("depth", defaultMaxDepth, "the maximum number of links deep to traverse")
	flag.Parse()

	links, err := CrawlWebpage(*urlFlag, *maxDepth)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	fmt.Println("Links")
	fmt.Println("-----")
	for i, l := range links {
		fmt.Printf("%03d. %s\n", i+1, l)
	}
	fmt.Println()
}
