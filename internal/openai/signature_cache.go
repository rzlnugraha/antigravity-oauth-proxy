package openai

import (
	"sync"
	"time"
)

type signatureEntry struct {
	signature string
	expiresAt time.Time
}

var (
	sigMu    sync.RWMutex
	sigCache = make(map[string]signatureEntry)
)

// StoreThoughtSignature stores a thought_signature keyed by tool call ID.
func StoreThoughtSignature(callID, signature string) {
	if callID == "" || signature == "" {
		return
	}
	sigMu.Lock()
	defer sigMu.Unlock()

	now := time.Now()
	// Periodic cleanup when cache gets large
	if len(sigCache) > 1000 {
		for k, v := range sigCache {
			if now.After(v.expiresAt) {
				delete(sigCache, k)
			}
		}
	}

	sigCache[callID] = signatureEntry{
		signature: signature,
		expiresAt: now.Add(2 * time.Hour),
	}
}

// GetThoughtSignature retrieves the thought_signature for a tool call ID.
func GetThoughtSignature(callID string) string {
	if callID == "" {
		return ""
	}
	sigMu.RLock()
	entry, ok := sigCache[callID]
	sigMu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return ""
	}
	return entry.signature
}
