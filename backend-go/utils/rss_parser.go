package utils

import (
	"crypto/md5"
	"fmt"
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

// NewRSSParser creates a new RSS parser
func NewRSSParser() *RSSParser {
	return &RSSParser{
		parser: gofeed.NewParser(),
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

	var items []*FeedItem
	for _, item := range feed.Items {
		feedItem := &FeedItem{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.Link,
			Author:      item.Author.Name,
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
