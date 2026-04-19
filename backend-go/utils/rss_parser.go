package utils

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
)

// RSSParser wraps gofeed parser
type RSSParser struct {
	parser   *gofeed.Parser
	proxyURL string
	mu       sync.Mutex
}

// NewRSSParser creates a new RSS parser, optionally with HTTP proxy
func NewRSSParser(proxyURL ...string) *RSSParser {
	fp := gofeed.NewParser()

	var transport *http.Transport
	if len(proxyURL) > 0 && proxyURL[0] != "" {
		proxy, err := url.Parse(proxyURL[0])
		if err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	if transport == nil {
		// Explicitly nil proxy — without this, Go inherits HTTP_PROXY env var,
		// which can silently route all RSS traffic through a container-level proxy
		transport = &http.Transport{
			Proxy: nil,
		}
	}

	fp.Client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	pURL := ""
	if len(proxyURL) > 0 {
		pURL = proxyURL[0]
	}

	return &RSSParser{
		parser:   fp,
		proxyURL: pURL,
	}
}

// SetProxyURL updates the proxy at runtime without restarting
func (rp *RSSParser) SetProxyURL(proxyURL string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.proxyURL = proxyURL

	var transport *http.Transport
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	if transport == nil {
		transport = &http.Transport{
			Proxy: nil,
		}
	}

	rp.parser.Client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// GetProxyURL returns the current proxy URL
func (rp *RSSParser) GetProxyURL() string {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	return rp.proxyURL
}

// FeedItem represents a single feed item
type FeedItem struct {
	Title       string
	Description string
	URL         string
	Author      string
	PublishedAt *time.Time
	Content     string
	ImageURLs   []string
}

// ParseFeed parses an RSS/Atom feed and returns items
func (rp *RSSParser) ParseFeed(feedURL string) ([]*FeedItem, error) {
	feed, err := rp.parser.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	if feed == nil {
		return nil, fmt.Errorf("feed is nil")
	}

	// Extract feed-level author as fallback
	feedAuthor := ""
	if feed.Author != nil && feed.Author.Name != "" {
		feedAuthor = feed.Author.Name
	} else if feed.Title != "" {
		feedAuthor = feed.Title
	}

	var items []*FeedItem
	for _, item := range feed.Items {
		author := ""
		if item.Author != nil && item.Author.Name != "" {
			author = item.Author.Name
		} else if len(item.Authors) > 0 && item.Authors[0].Name != "" {
			author = item.Authors[0].Name
		} else {
			author = feedAuthor
		}
		feedItem := &FeedItem{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			Author:      author,
			Content:     item.Content,
		}

		// Extract published time
		if item.PublishedParsed != nil {
			feedItem.PublishedAt = item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			feedItem.PublishedAt = item.UpdatedParsed
		} else {
			now := time.Now()
			feedItem.PublishedAt = &now
		}

		items = append(items, feedItem)
	}

	return items, nil
}

// ExtractImageURLs extracts all image URLs from HTML content
func ExtractImageURLs(html string) []string {
	if html == "" {
		return nil
	}

	re := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(html, -1)

	seen := make(map[string]bool)
	var urls []string
	for _, match := range matches {
		if len(match) > 1 {
			imgURL := strings.TrimSpace(match[1])
			if imgURL != "" && !seen[imgURL] {
				seen[imgURL] = true
				urls = append(urls, imgURL)
			}
		}
	}
	return urls
}

// htmlConverter is a reusable HTML-to-Markdown converter
var htmlConverter = md.NewConverter("", true, nil)

// CleanContent converts HTML content to Markdown, preserving structure
func CleanContent(content string) string {
	if content == "" {
		return ""
	}

	// Convert HTML to Markdown (preserves headings, lists, links, emphasis, code blocks, etc.)
	markdown, err := htmlConverter.ConvertString(content)
	if err != nil {
		log.Printf("[CleanContent] HTML-to-Markdown failed, falling back to plain text: %v", err)
		// Fallback: strip tags
		re := regexp.MustCompile(`<[^>]*>`)
		markdown = re.ReplaceAllString(content, "")
	}

	// Clean up excessive blank lines (3+ consecutive newlines → 2)
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")
	markdown = strings.TrimSpace(markdown)

	// Truncate to 2500 runes (rune-safe for CJK) — LLM context window budget
	runes := []rune(markdown)
	if len(runes) > 2500 {
		markdown = string(runes[:2500])
	}

	return markdown
}

// NormalizeURL normalizes a URL for consistency
func NormalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

// GenerateContentHash generates a hash for content (URL or title+content)
func GenerateContentHash(url, title, content string) string {
	// Use URL as primary hash source
	if url != "" {
		return fmt.Sprintf("%x", md5.Sum([]byte(NormalizeURL(url))))
	}

	// Fallback to title + content hash
	combined := title + "|" + content
	return fmt.Sprintf("%x", md5.Sum([]byte(combined)))
}

// SanitizeFeedItem cleans and normalizes a feed item
func SanitizeFeedItem(item *FeedItem) *FeedItem {
	item.URL = NormalizeURL(item.URL)
	item.Title = strings.TrimSpace(item.Title)

	rawHTML := item.Content
	if rawHTML == "" {
		rawHTML = item.Description
	}
	// Image extraction MUST happen before CleanContent — HTML-to-Markdown strips all <img> tags
	item.ImageURLs = ExtractImageURLs(rawHTML)

	item.Description = CleanContent(item.Description)
	if item.Content != "" {
		item.Content = CleanContent(item.Content)
	} else {
		item.Content = item.Description
	}

	return item
}
