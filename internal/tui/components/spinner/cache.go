package spinner

import (
	"fmt"
	"sync"

	"github.com/zeebo/xxh3"
)

// cache stores pre-rendered animation frames.
type cache struct {
	initialFrames  [][]string
	cyclingFrames  [][]string
	width          int
	labelWidth     int
	label          []string
	ellipsisFrames []string
}

var (
	cacheMap   = make(map[string]*cache)
	cacheMutex sync.RWMutex
)

// cacheKey creates a hash key from settings for cache lookup.
func cacheKey(s Settings) string {
	h := xxh3.New()
	fmt.Fprintf(h, "%d-%s-%v-%v-%v-%t",
		s.Size, s.Label, s.LabelColor, s.GradColorA, s.GradColorB, s.CycleColors)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// getCache retrieves cached animation data if it exists.
func getCache(key string) (*cache, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	c, ok := cacheMap[key]
	return c, ok
}

// setCache stores animation data in the cache.
func setCache(key string, c *cache) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheMap[key] = c
}
