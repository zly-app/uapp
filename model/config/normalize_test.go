package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 内部 viper 用自定义 KeyDelimiter 是为了让 apollo 的命名空间 key（frame.yaml）
// 不被拆成层级；但交给 zapp 之前必须换回默认 `.` 分隔符，
// 否则 UnmarshalKey("plugins.pprof") 这类点路径读不到配置。
func TestResolveNormalizesKeyDelimiterForZapp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "default.yaml")
	content := `
plugins:
  pprof:
    Enable: true
    Bind: 127.0.0.1:16060
services:
  grpc:
    hello:
      Bind: 127.0.0.1:0
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写配置失败: %v", err)
	}

	r, err := Resolve(Source{DefaultFiles: []string{path}})
	if err != nil {
		t.Fatalf("Resolve 失败: %v", err)
	}

	var pconf struct {
		Enable bool
		Bind   string
	}
	if err := r.Viper.UnmarshalKey("plugins.pprof", &pconf); err != nil {
		t.Fatalf("读取 plugins.pprof 失败: %v", err)
	}
	if !pconf.Enable || pconf.Bind != "127.0.0.1:16060" {
		t.Fatalf("plugins.pprof 解析不符: %+v", pconf)
	}

	var sconf struct {
		Bind string
	}
	if err := r.Viper.UnmarshalKey("services.grpc.hello", &sconf); err != nil {
		t.Fatalf("读取 services.grpc.hello 失败: %v", err)
	}
	if sconf.Bind != "127.0.0.1:0" {
		t.Fatalf("services.grpc.hello.Bind 解析不符: %q", sconf.Bind)
	}
}

// apollo 命名空间 key 在归一化后仍要能整体取出
func TestNormalizeKeepsNamespaceKeys(t *testing.T) {
	vi := newViper()
	vi.Set("frame.yaml", "Name: demo")

	out := normalize(vi)
	if got := out.GetString("frame.yaml"); got != "Name: demo" {
		t.Fatalf("命名空间 key 丢失: %q", got)
	}
	if got := vi.GetString("frame.yaml"); got != "Name: demo" {
		t.Fatalf("内部 viper 也被破坏: %q", got)
	}
}
