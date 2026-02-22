package mcp

import "sync"

// ConsoleEntry represents a single console log entry from the webview.
type ConsoleEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// ConsoleBuffer is a thread-safe ring buffer for console log entries.
type ConsoleBuffer struct {
	mu      sync.Mutex
	entries []ConsoleEntry
	maxSize int
}

// NewConsoleBuffer creates a new console buffer with the given maximum size.
func NewConsoleBuffer(maxSize int) *ConsoleBuffer {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &ConsoleBuffer{
		entries: make([]ConsoleEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends an entry to the buffer. If the buffer is full, the oldest
// entry is dropped.
func (b *ConsoleBuffer) Add(entry ConsoleEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) >= b.maxSize {
		// Drop the oldest entry
		copy(b.entries, b.entries[1:])
		b.entries = b.entries[:len(b.entries)-1]
	}
	b.entries = append(b.entries, entry)
}

// Get returns the last n entries, optionally filtered by level.
// If level is "all" or empty, all entries are returned.
// If n is 0 or negative, all matching entries are returned.
func (b *ConsoleBuffer) Get(n int, level string) []ConsoleEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	if level == "" || level == "all" {
		// No filtering
		if n <= 0 || n >= len(b.entries) {
			result := make([]ConsoleEntry, len(b.entries))
			copy(result, b.entries)
			return result
		}
		start := len(b.entries) - n
		result := make([]ConsoleEntry, n)
		copy(result, b.entries[start:])
		return result
	}

	// Filter by level
	var filtered []ConsoleEntry
	for _, e := range b.entries {
		if e.Level == level {
			filtered = append(filtered, e)
		}
	}

	if n <= 0 || n >= len(filtered) {
		return filtered
	}

	start := len(filtered) - n
	return filtered[start:]
}

// Clear removes all entries from the buffer.
func (b *ConsoleBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = b.entries[:0]
}
