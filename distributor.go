package heimdall

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// APIKey represents a single API key with its quota information
type APIKey struct {
	Secret       string
	Name         string
	MaxRequests  uint32        // Maximum requests allowed for this key
	RequestsUsed uint32        // Current count of requests used
	ResetTime    time.Time     // When the quota resets
	QuotaPeriod  time.Duration // Period after which the quota resets
	mu           sync.Mutex
}

// Available returns the number of requests still available for this key
func (k *APIKey) Available() uint32 {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if quota should be reset
	if time.Now().After(k.ResetTime) {
		k.RequestsUsed = 0
		k.ResetTime = time.Now().Add(k.QuotaPeriod)
	}

	if k.RequestsUsed >= k.MaxRequests {
		return 0
	}

	return k.MaxRequests - k.RequestsUsed
}

// UseRequest increments the usage counter for this key
// Returns true if the request was allowed, false if quota exceeded
func (k *APIKey) UseRequest() bool {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if quota should be reset
	if time.Now().After(k.ResetTime) {
		k.RequestsUsed = 0
		k.ResetTime = time.Now().Add(k.QuotaPeriod)
	}

	if k.RequestsUsed >= k.MaxRequests {
		return false
	}

	k.RequestsUsed++
	return true
}

// KeyDistributor manages multiple API keys and distributes requests among them
type KeyDistributor struct {
	keys      []*APIKey
	lastIndex int
	mu        sync.Mutex
}

// KeyConfig holds configuration for an individual API key
type KeyConfig struct {
	Key         string
	MaxRequests uint32
	QuotaPeriod time.Duration
}

// NewKeyDistributor creates a new key distributor with the given keys and their quotas
func NewKeyDistributor(keyConfigs []KeyConfig) (*KeyDistributor, error) {
	if len(keyConfigs) == 0 {
		return nil, errors.New("at least one API key is required")
	}

	keys := make([]*APIKey, len(keyConfigs))
	for i, config := range keyConfigs {
		if config.MaxRequests == 0 {
			return nil, errors.New("MaxRequests must be greater than zero")
		}

		// Default period is 1 hour if not specified
		period := config.QuotaPeriod
		if period == 0 {
			period = time.Hour
		}

		keys[i] = &APIKey{
			Secret:      config.Key,
			MaxRequests: config.MaxRequests,
			QuotaPeriod: period,
			ResetTime:   time.Now().Add(period),
		}
	}

	return &KeyDistributor{
		keys:      keys,
		lastIndex: -1,
	}, nil
}

// GetNextKey returns the next available API key in a round-robin fashion,
// prioritizing keys with more available requests
func (d *KeyDistributor) GetNextKey() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.keys) == 0 {
		return "", errors.New("no API keys available")
	}

	// Step 1: Check if any keys have available requests
	allExhausted := true
	for _, key := range d.keys {
		if key.Available() > 0 {
			allExhausted = false
			break
		}
	}

	if allExhausted {
		return "", errors.New("all API keys have exhausted their quota")
	}

	// Step 2: Start from the next key after the last used one (round-robin)
	startIndex := (d.lastIndex + 1) % len(d.keys)

	// First attempt: try to find an available key in round-robin order
	for i := 0; i < len(d.keys); i++ {
		index := (startIndex + i) % len(d.keys)
		if d.keys[index].Available() > 0 {
			// Found an available key, use it
			if d.keys[index].UseRequest() {
				d.lastIndex = index
				return d.keys[index].Secret, nil
			}
		}
	}

	// If we get here, something went wrong - we should have found an available key
	return "", errors.New("failed to allocate an API key")
}

// GetOptimalKey returns the key with the most available requests
// This can be used when strict round-robin isn't required
func (d *KeyDistributor) GetOptimalKey() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.keys) == 0 {
		return "", errors.New("no API keys available")
	}

	// Create a copy of keys for sorting
	type keyWithAvailability struct {
		key       *APIKey
		available uint32
	}

	keysCopy := make([]keyWithAvailability, len(d.keys))
	for i, key := range d.keys {
		keysCopy[i] = keyWithAvailability{
			key:       key,
			available: key.Available(),
		}
	}

	// Sort by available requests (descending)
	sort.Slice(keysCopy, func(i, j int) bool {
		return keysCopy[i].available > keysCopy[j].available
	})

	// Check if the best key has any availability
	if keysCopy[0].available == 0 {
		return "", errors.New("all API keys have exhausted their quota")
	}

	// Use the key with the most available requests
	bestKey := keysCopy[0].key
	if bestKey.UseRequest() {
		// Update lastIndex for this key for future reference
		for i, key := range d.keys {
			if key == bestKey {
				d.lastIndex = i
				break
			}
		}
		return bestKey.Secret, nil
	}

	return "", errors.New("failed to allocate an API key")
}

// GetUsage returns the current usage information for all keys
func (d *KeyDistributor) GetUsage() map[string]struct {
	Used      uint32
	Available uint32
	MaxQuota  uint32
	ResetAt   time.Time
} {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := make(map[string]struct {
		Used      uint32
		Available uint32
		MaxQuota  uint32
		ResetAt   time.Time
	}, len(d.keys))

	for _, key := range d.keys {
		key.mu.Lock()
		result[key.Secret] = struct {
			Used      uint32
			Available uint32
			MaxQuota  uint32
			ResetAt   time.Time
		}{
			Used:      key.RequestsUsed,
			Available: key.MaxRequests - key.RequestsUsed,
			MaxQuota:  key.MaxRequests,
			ResetAt:   key.ResetTime,
		}
		key.mu.Unlock()
	}

	return result
}

// ResetUsage resets the usage counters for all keys
func (d *KeyDistributor) ResetUsage() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, key := range d.keys {
		key.mu.Lock()
		key.RequestsUsed = 0
		key.ResetTime = time.Now().Add(key.QuotaPeriod)
		key.mu.Unlock()
	}
}
