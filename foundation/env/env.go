// Package env 提供环境变量读取。
//
// 这是底座机制：只做「按 key 取值 + 类型转换 + 默认值」，不含任何业务语义，
// 也不持有全局状态。配置字段的含义与默认值由 model/config 决定。
package env

import (
	"os"
	"strings"

	"github.com/spf13/cast"
)

// String 读取字符串；未设置或为空时返回 def。
func String(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Bool 读取布尔值；未设置或无法解析时返回 def。
func Bool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	b, err := cast.ToBoolE(v)
	if err != nil {
		return def
	}
	return b
}

// List 读取逗号分隔的列表，去掉空白与空项；未设置时返回 nil。
func List(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}

	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
