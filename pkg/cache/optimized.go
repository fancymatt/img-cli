// Package cache provides optimized caching with memory efficiency improvements.
package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"img-cli/pkg/logger"
	"img-cli/pkg/models"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// OptimizedCache provides thread-safe, memory-efficient caching
type OptimizedCache struct {
	cacheDir string
	ttl      time.Duration
	mu       sync.RWMutex
	index    map[string]*IndexEntry // In-memory index for fast lookups
}

// IndexEntry represents cached metadata without loading full data
type IndexEntry struct {
	Key       string    `json:"key"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"file_path"`
	FileHash  string    `json:"file_hash"`
	Size      int64     `json:"size"`
}

// NewOptimizedCache creates a new optimized cache instance
func NewOptimizedCache(cacheDir string, ttl time.Duration) *OptimizedCache {
	if cacheDir == "" {
		cacheDir = ".cache/analyses"
	}
	if ttl == 0 {
		ttl = 24 * time.Hour * 7 // Default 7 days
	}

	os.MkdirAll(cacheDir, 0755)

	cache := &OptimizedCache{
		cacheDir: cacheDir,
		ttl:      ttl,
		index:    make(map[string]*IndexEntry),
	}

	// Build index on initialization
	cache.buildIndex()

	return cache
}

// buildIndex scans cache directory and builds in-memory index
func (c *OptimizedCache) buildIndex() {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Debug("Building cache index", "dir", c.cacheDir)

	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		logger.Warn("Failed to read cache directory", "error", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(c.cacheDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Try to load just the metadata
		var meta CacheEntry
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// Check if expired
		if time.Since(meta.Timestamp) > c.ttl {
			os.Remove(path) // Clean up expired entries
			continue
		}

		// Add to index
		c.index[meta.Key] = &IndexEntry{
			Key:       meta.Key,
			Type:      meta.Type,
			Timestamp: meta.Timestamp,
			FilePath:  meta.FilePath,
			FileHash:  meta.FileHash,
			Size:      info.Size(),
		}
	}

	logger.Info("Cache index built", "entries", len(c.index))
}

// GetOutfitAnalysis retrieves outfit analysis from cache with type safety
func (c *OptimizedCache) GetOutfitAnalysis(filePath string) (*models.OutfitAnalysis, bool) {
	key := c.generateKey("outfit", filePath)

	c.mu.RLock()
	entry, exists := c.index[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check expiry
	if time.Since(entry.Timestamp) > c.ttl {
		c.evict(key)
		return nil, false
	}

	// Load full data
	cachePath := filepath.Join(c.cacheDir, key+".json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cacheEntry CacheEntry
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		return nil, false
	}

	// Verify file hash if needed
	currentHash, err := c.getFileHash(filePath)
	if err == nil && currentHash != entry.FileHash {
		c.evict(key)
		return nil, false
	}

	// Parse the outfit analysis
	var analysis models.OutfitAnalysis
	if err := json.Unmarshal(cacheEntry.Data, &analysis); err != nil {
		return nil, false
	}

	logger.Debug("Cache hit", "type", "outfit", "key", key)
	return &analysis, true
}

// SetOutfitAnalysis stores outfit analysis in cache
func (c *OptimizedCache) SetOutfitAnalysis(filePath string, analysis *models.OutfitAnalysis) error {
	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	return c.Set("outfit", filePath, json.RawMessage(data))
}

// GetVisualStyleAnalysis retrieves visual style analysis from cache
func (c *OptimizedCache) GetVisualStyleAnalysis(filePath string) (*models.VisualStyleAnalysis, bool) {
	key := c.generateKey("visual_style", filePath)

	c.mu.RLock()
	entry, exists := c.index[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Since(entry.Timestamp) > c.ttl {
		c.evict(key)
		return nil, false
	}

	cachePath := filepath.Join(c.cacheDir, key+".json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cacheEntry CacheEntry
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		return nil, false
	}

	var analysis models.VisualStyleAnalysis
	if err := json.Unmarshal(cacheEntry.Data, &analysis); err != nil {
		return nil, false
	}

	logger.Debug("Cache hit", "type", "visual_style", "key", key)
	return &analysis, true
}

// SetVisualStyleAnalysis stores visual style analysis in cache
func (c *OptimizedCache) SetVisualStyleAnalysis(filePath string, analysis *models.VisualStyleAnalysis) error {
	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	return c.Set("visual_style", filePath, json.RawMessage(data))
}

// evict removes an entry from cache
func (c *OptimizedCache) evict(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.index, key)
	cachePath := filepath.Join(c.cacheDir, key+".json")
	os.Remove(cachePath)

	logger.Debug("Cache entry evicted", "key", key)
}

// GetStats returns cache statistics with efficient calculation
func (c *OptimizedCache) GetStats() (*models.CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &models.CacheStats{
		TotalEntries:  len(c.index),
		EntriesByType: make(map[string]int),
	}

	var totalSize int64
	var oldest, newest time.Time

	for _, entry := range c.index {
		// Count by type
		stats.EntriesByType[entry.Type]++

		// Track size
		totalSize += entry.Size

		// Track oldest/newest
		if oldest.IsZero() || entry.Timestamp.Before(oldest) {
			oldest = entry.Timestamp
		}
		if newest.IsZero() || entry.Timestamp.After(newest) {
			newest = entry.Timestamp
		}
	}

	stats.TotalSize = totalSize
	stats.OldestEntry = oldest
	stats.NewestEntry = newest

	return stats, nil
}

// ClearType removes all cache entries of a specific type
func (c *OptimizedCache) ClearType(cacheType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	keysToDelete := []string{}
	for key, entry := range c.index {
		if entry.Type == cacheType {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(c.index, key)
		cachePath := filepath.Join(c.cacheDir, key+".json")
		os.Remove(cachePath)
	}

	logger.Info("Cache type cleared", "type", cacheType, "entries", len(keysToDelete))
	return nil
}

// Cleanup removes expired entries
func (c *OptimizedCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	expired := []string{}
	now := time.Now()

	for key, entry := range c.index {
		if now.Sub(entry.Timestamp) > c.ttl {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(c.index, key)
		cachePath := filepath.Join(c.cacheDir, key+".json")
		os.Remove(cachePath)
	}

	if len(expired) > 0 {
		logger.Info("Cache cleanup completed", "expired", len(expired))
	}
}

// StartCleanupRoutine starts a background cleanup routine
func (c *OptimizedCache) StartCleanupRoutine(interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Hour
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.Cleanup()
		}
	}()

	logger.Info("Cache cleanup routine started", "interval", interval)
}

// generateKey generates a cache key from analysis type and file path
func (c *OptimizedCache) generateKey(analysisType, filePath string) string {
	baseName := filepath.Base(filePath)
	cleanName := strings.ReplaceAll(baseName, " ", "_")
	return analysisType + "_" + cleanName
}

// getFileHash calculates the hash of a file
func (c *OptimizedCache) getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	// For large files, use size + modification time
	if info.Size() > 10*1024*1024 {
		hashStr := strings.Join([]string{
			"size", string(rune(info.Size())),
			"mod", string(rune(info.ModTime().Unix())),
		}, "_")
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

// Set stores data in the cache (delegates to parent Cache.Set)
func (c *OptimizedCache) Set(analysisType, filePath string, data json.RawMessage) error {
	key := c.generateKey(analysisType, filePath)
	cachePath := filepath.Join(c.cacheDir, key+".json")

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

	// Update index
	c.mu.Lock()
	c.index[key] = &IndexEntry{
		Key:       key,
		Type:      analysisType,
		Timestamp: entry.Timestamp,
		FilePath:  absPath,
		FileHash:  fileHash,
		Size:      int64(len(jsonData)),
	}
	c.mu.Unlock()

	return os.WriteFile(cachePath, jsonData, 0644)
}