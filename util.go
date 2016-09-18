package lwlog

import (
	"runtime"
	"strings"
	"sync"
)

var (
	loggingFunctions     map[uintptr]bool
	loggingFunctionsLock sync.RWMutex
)

// IsLoggingFunction returns true if the given function is known to be part of
// the logging infrastructure. It does this by examining f.Name() to determine
// if it is one of Debugf, Infof or Logf.
func IsLoggingFunction(f *runtime.Func) bool {
	var isLogger, known bool

	// we'll use the function's entrypoint address as the key in our cache
	pc := f.Entry()

	// test whether we've already cached the answer for this function
	loggingFunctionsLock.RLock()
	if loggingFunctions != nil {
		isLogger, known = loggingFunctions[pc]
	}
	loggingFunctionsLock.RUnlock()
	if known {
		return isLogger
	}

	// no cache entry — so determine whether we consider this a logging
	// function
	fn := f.Name()
	isLogger = strings.HasSuffix(fn, ".Debugf") ||
		strings.HasSuffix(fn, ".Infof") ||
		strings.HasSuffix(fn, ".Logf")

	// update the cache (NB: it's possible that another goroutine could
	// also have passed into or through the above lock in the meantime,
	// but that's fine — it will just write the same result into the map).
	loggingFunctionsLock.Lock()
	if loggingFunctions == nil {
		loggingFunctions = make(map[uintptr]bool)
	}
	loggingFunctions[pc] = isLogger
	loggingFunctionsLock.Unlock()

	return isLogger
}
