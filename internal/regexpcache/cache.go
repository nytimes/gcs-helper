package regexpcache

import (
	"regexp"
	"sync"
)

var cache sync.Map

// MustCompile works like regexp.MustCompile, but before compiling the regexp,
// it checks if it's available in the cache.
func MustCompile(str string) *regexp.Regexp {
	v, ok := cache.Load(str)
	if !ok {
		v = regexp.MustCompile(str)
		cache.Store(str, v)
	}
	return v.(*regexp.Regexp)
}
