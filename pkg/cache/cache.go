package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"img-cli/pkg/models"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cache struct {
	cacheDir string
	ttl      time.Duration
}

type CacheEntry struct {
	Key       string          `json:"key"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	FilePath  string          `json:"file_path"`
	FileHash  string          `json:"file_hash"`
	Data      json.RawMessage `json:"data"`
}

func NewCache(cacheDir string, ttl time.Duration) *Cache {
	if cacheDir == "" {
		cacheDir = ".cache/analyses"
	}
	if ttl == 0 {
		ttl = 24 * time.Hour * 7 // Default 7 days
	}

	os.MkdirAll(cacheDir, 0755)

	return &Cache{
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

// NewCacheForType creates a cache instance for a specific analysis type
func NewCacheForType(analysisType string, ttl time.Duration) *Cache {
	var cacheDir string

	switch analysisType {
	case "outfit":
		cacheDir = "outfits/.cache"
	case "visual_style", "art_style":
		cacheDir = "styles/.cache"
	default:
		cacheDir = ".cache/analyses"
	}

	if ttl == 0 {
		ttl = 24 * time.Hour * 7 // Default 7 days
	}

	os.MkdirAll(cacheDir, 0755)

	return &Cache{
		cacheDir: cacheDir,
		ttl:      ttl,
	}
}

func (c *Cache) generateKey(analysisType, filePath string) string {
	// Use just the filename (base name) for the key, not the full path
	// This allows the cache to work even if files are moved to different directories
	baseName := filepath.Base(filePath)
	// Clean the filename to be filesystem-safe
	cleanName := strings.ReplaceAll(baseName, " ", "_")
	return fmt.Sprintf("%s_%s", analysisType, cleanName)
}

func (c *Cache) getFileHash(filePath string) (string, error) {
	// Calculate hash based on actual file content, not path
	// This ensures the same file has the same hash regardless of location
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get file info for size check
	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	// For large files (>10MB), use size + modification time for performance
	if info.Size() > 10*1024*1024 {
		hashStr := fmt.Sprintf("size_%d_mod_%d", info.Size(), info.ModTime().Unix())
		h := md5.New()
		h.Write([]byte(hashStr))
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	// For smaller files, hash the actual content
	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *Cache) Get(analysisType, filePath string) (json.RawMessage, bool) {
	key := c.generateKey(analysisType, filePath)
	cachePath := filepath.Join(c.cacheDir, key+".json")

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// IMPORTANT: Always use cached version if it exists
	// This allows manual edits to be preserved
	// We don't check TTL expiration or file hash changes
	// The cache is based purely on filename, not path or content

	return entry.Data, true
}

func (c *Cache) Set(analysisType, filePath string, data json.RawMessage) error {
	key := c.generateKey(analysisType, filePath)
	cachePath := filepath.Join(c.cacheDir, key+".json")

	// IMPORTANT: Never overwrite existing cache files
	// This preserves manual edits made to cache files
	if _, err := os.Stat(cachePath); err == nil {
		// Cache file already exists, don't overwrite it
		return nil
	}

	absPath, _ := filepath.Abs(filePath)
	fileHash, err := c.getFileHash(filePath)
	if err != nil {
		fileHash = ""
	}

	entry := CacheEntry{
		Key:       key,
		Type:      analysisType,
		Timestamp: time.Now(),
		FilePath:  absPath,
		FileHash:  fileHash,
		Data:      data,
	}

	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, jsonData, 0644)
}

func (c *Cache) Clear() error {
	return os.RemoveAll(c.cacheDir)
}

func (c *Cache) ClearType(analysisType string) error {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(c.cacheDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if entry.Type == analysisType {
			os.Remove(filePath)
		}
	}

	return nil
}

func (c *Cache) Stats() (map[string]interface{}, error) {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_entries": len(files),
		"cache_dir":     c.cacheDir,
		"ttl_hours":     c.ttl.Hours(),
		"by_type":       make(map[string]int),
	}

	byType := stats["by_type"].(map[string]int)

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(c.cacheDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		byType[entry.Type]++
	}

	return stats, nil
}

// GetStats returns cache statistics in the models.CacheStats format
func (c *Cache) GetStats() (*models.CacheStats, error) {
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return nil, err
	}

	stats := &models.CacheStats{
		TotalEntries:  0,
		EntriesByType: make(map[string]int),
		TotalSize:     0,
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		stats.TotalEntries++

		info, err := file.Info()
		if err == nil {
			stats.TotalSize += info.Size()
		}

		filePath := filepath.Join(c.cacheDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		stats.EntriesByType[entry.Type]++

		// Track oldest/newest
		if stats.OldestEntry.IsZero() || entry.Timestamp.Before(stats.OldestEntry) {
			stats.OldestEntry = entry.Timestamp
		}
		if stats.NewestEntry.IsZero() || entry.Timestamp.After(stats.NewestEntry) {
			stats.NewestEntry = entry.Timestamp
		}
	}

	return stats, nil
}