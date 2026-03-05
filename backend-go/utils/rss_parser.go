package utils

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// RSSParser wraps gofeed parser
type RSSParser struct {
	parser *gofeed.Parser
}

// NewRSSParser creates a new RSS parser, optionally with HTTP proxy
func NewRSSParser(proxyURL ...string) *RSSParser {
	fp := gofeed.NewParser()

	if len(proxyURL) > 0 && proxyURL[0] != "" {
		proxy, err := url.Parse(proxyURL[0])
		if err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
			fp.Client = &http.Client{
				Transport: transport,
				Timeout:   30 * time.Second,
			}
		}
	}

	return &RSSParser{
		parser: fp,
	}
}

// FeedItem represents a single feed item
type FeedItem struct {
	Title       string
	Description string
	URL         string
	Author      string
	PublishedAt *time.Time
	Content     string
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

// CleanContent removes HTML tags and unnecessary whitespace
func CleanContent(content string) string {
	if content == "" {
		return ""
	}

	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Decode HTML entities
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")

	// Remove extra whitespace
	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// Truncate if too long
	if len(content) > 5000 {
		content = content[:5000]
	}

	return content
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
	item.Description = CleanContent(item.Description)
	if item.Content != "" {
		item.Content = CleanContent(item.Content)
	} else {
		item.Content = item.Description
	}

	return item
}
