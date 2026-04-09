package repositories

// FetchWindow trims results and reports whether more rows exist.
// Callers should query with limit+1 and pass the original limit for correct trimming and hasMore.
func FetchWindow[T any](items []T, limit int) (trimmed []T, hasMore bool) {
	if len(items) > limit {
		return items[:limit], true
	}
	return items, false
}
