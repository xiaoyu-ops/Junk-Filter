package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spaolacci/murmur3"
	"github.com/junkfilter/backend-go/repositories"
	"github.com/junkfilter/backend-go/utils"
)

// BloomFilter simple in-memory bloom filter implementation
type BloomFilter struct {
	bits []byte
	size uint64
}

// NewBloomFilter creates a new bloom filter with 7-day window
func NewBloomFilter(expectedItems int) *BloomFilter {
	// Estimate bit size: -1/(ln(2)^2) * n * ln(p)
	// For 1 million items and 0.1% error rate
	size := uint64(expectedItems * 1000 / 100) // ~10M bits
	return &BloomFilter{
		bits: make([]byte, size/8+1),
		size: size,
	}
}

// Add adds an item to the bloom filter
func (bf *BloomFilter) Add(item string) {
	h1 := murmur3.Sum32([]byte(item))
	h2 := murmur3.Sum32([]byte(item + "2"))

	pos1 := uint64(h1) % bf.size
	pos2 := uint64(h2) % bf.size

	bf.bits[pos1/8] |= 1 << (pos1 % 8)
	bf.bits[pos2/8] |= 1 << (pos2 % 8)
}

// Contains checks if an item might exist in the bloom filter
func (bf *BloomFilter) Contains(item string) bool {
	h1 := murmur3.Sum32([]byte(item))
	h2 := murmur3.Sum32([]byte(item + "2"))

	pos1 := uint64(h1) % bf.size
	pos2 := uint64(h2) % bf.size

	return (bf.bits[pos1/8]&(1<<(pos1%8)) != 0) && (bf.bits[pos2/8]&(1<<(pos2%8)) != 0)
}

// DedupService handles three-layer deduplication
type DedupService struct {
	bloomFilter *BloomFilter
	redis       *redis.Client
	contentRepo *repositories.ContentRepository
}

// NewDedupService creates a new dedup service
func NewDedupService(redis *redis.Client, contentRepo *repositories.ContentRepository) *DedupService {
	return &DedupService{
		bloomFilter: NewBloomFilter(1000000),
		redis:       redis,
		contentRepo: contentRepo,
	}
}

// InitializeBloomFilter loads existing URLs from DB into bloom filter
func (ds *DedupService) InitializeBloomFilter(ctx context.Context) error {
	// Load last 7 days of content
	// This is a simplified version - in production, you'd query all URLs from last 7 days
	// For now, we'll just load URLs from Redis keys

	keys, err := ds.redis.Keys(ctx, "dedup:url:*").Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, key := range keys {
		// Extract URL from key format "dedup:url:{hash}"
		if len(key) > 11 {
			urlHash := key[11:]
			ds.bloomFilter.Add(urlHash)
		}
	}

	return nil
}

// IsDuplicate checks if URL is duplicate using three-layer dedup
func (ds *DedupService) IsDuplicate(ctx context.Context, url, contentHash string) (bool, error) {
	// L1: Bloom Filter (fast rejection)
	if ds.bloomFilter.Contains(url) {
		// L2: Redis Set (exact check)
		redisKey := fmt.Sprintf("dedup:url:%s", url)
		exists, err := ds.redis.Exists(ctx, redisKey).Result()
		if err != nil && err != redis.Nil {
			return false, err
		}

		if exists > 0 {
			return true, nil // Confirmed duplicate
		}
	}

	// L3: Database constraint will catch remaining duplicates
	// This is handled by UNIQUE constraints on original_url and content_hash

	return false, nil
}

// MarkAsSeen marks a URL/hash as seen
func (ds *DedupService) MarkAsSeen(ctx context.Context, url, contentHash string) error {
	// Add to bloom filter
	ds.bloomFilter.Add(url)

	// Add to Redis with 7-day TTL
	redisKey := fmt.Sprintf("dedup:url:%s", url)
	return ds.redis.Set(ctx, redisKey, contentHash, 7*24*time.Hour).Err()
}

// CheckContentHash checks if content hash exists
func (ds *DedupService) CheckContentHash(ctx context.Context, hash string) (bool, error) {
	redisKey := fmt.Sprintf("dedup:hash:%s", hash)
	exists, err := ds.redis.Exists(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return exists > 0, nil
}

// MarkHashAsSeen marks a content hash as seen
func (ds *DedupService) MarkHashAsSeen(ctx context.Context, hash string) error {
	redisKey := fmt.Sprintf("dedup:hash:%s", hash)
	return ds.redis.Set(ctx, redisKey, "1", 7*24*time.Hour).Err()
}

// ValidateContent generates hash and checks for duplicates
func (ds *DedupService) ValidateContent(ctx context.Context, url, title, content string) (string, bool, error) {
	// Generate content hash
	contentHash := utils.GenerateContentHash(url, title, content)

	// Check if URL is duplicate
	isDup, err := ds.IsDuplicate(ctx, url, contentHash)
	if err != nil {
		return contentHash, false, err
	}

	if isDup {
		return contentHash, true, nil
	}

	// Check if content hash is duplicate (different URL, same content)
	isHashDup, err := ds.CheckContentHash(ctx, contentHash)
	if err != nil {
		return contentHash, false, err
	}

	return contentHash, isHashDup, nil
}
