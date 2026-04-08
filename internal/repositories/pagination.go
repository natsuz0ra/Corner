package repositories

// FetchWindow 截断查询结果并判断是否有更多数据。
// 调用方应使用 limit+1 查询，传入原始 limit 以获得正确的截断与 hasMore 判定。
func FetchWindow[T any](items []T, limit int) (trimmed []T, hasMore bool) {
	if len(items) > limit {
		return items[:limit], true
	}
	return items, false
}
