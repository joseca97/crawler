package crawler

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func extractLinks(baseURL string, body io.Reader) []string {
	var links []string

	doc, err := html.Parse(body)
	if err != nil {
		return nil
	}

	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					cleaned := cleanLink(attr.Val)
					if cleaned != "" {
						absoluteURL := resolveAbsoluteURL(baseURL, cleaned)
						if absoluteURL != "" {
							links = append(links, absoluteURL)
						}
					}
					break
				}
			}
		}

		// for c := n.FirstChild; c != nil; c = n.NextSibling {
		// 	visit(c)
		// }
		if n.FirstChild != nil {
			visit(n.FirstChild)
		}
		if n.NextSibling != nil {
			visit(n.NextSibling)
		}
	}

	visit(doc)
	return links
}

func cleanLink(href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
		return ""
	}
	return href
}

func resolveAbsoluteURL(baseURL, targetURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	target, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	// ResolveReference handles all edge cases perfectly:
	// - If targetURL is already "https://google.com", it leaves it alone.
	// - If targetURL is "/docs", it turns it into "https://golang.org/docs".
	// - If targetURL is "docs/intro.html", it appends it relative to the current directory path.
	resolved := base.ResolveReference(target)

	// Strip out URL fragments (like "#it-works") so we don't crawl the exact same page twice
	resolved.Fragment = ""

	// We only want to crawl HTTP and HTTPS links
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}
