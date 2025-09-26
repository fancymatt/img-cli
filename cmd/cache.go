package cmd

import (
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"img-cli/pkg/workflow"

	"github.com/spf13/cobra"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache <action>",
	Short: "Manage the analysis cache",
	Long: `Manage the cache for analysis results.

Available actions:
  stats              - Show cache statistics
  clear              - Clear all cache entries
  clear-outfit       - Clear outfit analysis cache
  clear-visual_style - Clear visual style cache
  clear-art_style    - Clear art style cache`,
	Args: cobra.ExactArgs(1),
	RunE: runCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}

func runCache(cmd *cobra.Command, args []string) error {
	action := args[0]
	orchestrator := workflow.NewOrchestrator(apiKey)

	switch action {
	case "stats":
		// Get stats from all caches
		totalEntries := 0
		totalSize := int64(0)
		entriesByType := make(map[string]int)

		for _, cacheType := range []string{"outfit", "visual_style", "art_style"} {
			cache := orchestrator.GetCacheForType(cacheType)
			stats, err := cache.GetStats()
			if err != nil {
				continue
			}
			totalEntries += stats.TotalEntries
			totalSize += stats.TotalSize
			for typ, count := range stats.EntriesByType {
				entriesByType[typ] += count
			}
		}

		fmt.Println("Cache Statistics (All Locations):")
		fmt.Printf("  Total entries: %d\n", totalEntries)
		fmt.Printf("  Total size: %.2f MB\n", float64(totalSize)/1024/1024)
		fmt.Println("\nCache locations:")
		fmt.Println("  Outfit cache: outfits/.cache")
		fmt.Println("  Style caches: styles/.cache")

		if len(entriesByType) > 0 {
			fmt.Println("\nEntries by type:")
			for typ, count := range entriesByType {
				fmt.Printf("    %s: %d\n", typ, count)
			}
		}

		logger.Info("Cache stats retrieved",
			"entries", totalEntries,
			"size_mb", float64(totalSize)/1024/1024)

	case "clear":
		// Clear all caches
		for _, cacheType := range []string{"outfit", "visual_style", "art_style"} {
			cache := orchestrator.GetCacheForType(cacheType)
			if err := cache.Clear(); err != nil {
				logger.Warn("Failed to clear cache", "type", cacheType, "error", err)
			}
		}
		fmt.Println("✓ All caches cleared successfully")
		logger.Info("All caches cleared")

	case "clear-outfit":
		cache := orchestrator.GetCacheForType("outfit")
		if err := cache.ClearType("outfit"); err != nil {
			return errors.Wrap(err, errors.CacheError, "failed to clear outfit cache")
		}
		fmt.Println("✓ Outfit cache cleared successfully (outfits/.cache)")
		logger.Info("Outfit cache cleared")

	case "clear-visual_style":
		cache := orchestrator.GetCacheForType("visual_style")
		if err := cache.ClearType("visual_style"); err != nil {
			return errors.Wrap(err, errors.CacheError, "failed to clear visual style cache")
		}
		fmt.Println("✓ Visual style cache cleared successfully (styles/.cache)")
		logger.Info("Visual style cache cleared")

	case "clear-art_style":
		cache := orchestrator.GetCacheForType("art_style")
		if err := cache.ClearType("art_style"); err != nil {
			return errors.Wrap(err, errors.CacheError, "failed to clear art style cache")
		}
		fmt.Println("✓ Art style cache cleared successfully (styles/.cache)")
		logger.Info("Art style cache cleared")

	default:
		return errors.ErrInvalidInput("action", fmt.Sprintf("unknown action: %s", action))
	}

	return nil
}