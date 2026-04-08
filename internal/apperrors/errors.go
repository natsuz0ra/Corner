package apperrors

import "errors"

// 领域级哨兵错误，供仓储层返回、服务层/控制器层判断。
var (
	// ErrNotFound 表示请求的资源不存在。
	ErrNotFound = errors.New("not found")
	// ErrInvalidInput 表示输入参数不合法。
	ErrInvalidInput = errors.New("invalid input")
)
