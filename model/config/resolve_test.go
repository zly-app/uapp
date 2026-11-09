package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试配置失败: %v", err)
	}
	return path
}

func TestResolveWithoutApollo(t *testing.T) {
	dir := t.TempDir()
	file := writeConfigFile(t, dir, "default.yaml", "frame:\n  Debug: false\nmyconf:\n  A: 1\n")

	got, err := Resolve(Source{
		AppName:      "svc",
		DefaultFiles: []string{filepath.Join(dir, "missing.yaml"), file},
	})
	if err != nil {
		t.Fatalf("Resolve 失败: %v", err)
	}
	if got.ConfigFile != file {
		t.Errorf("ConfigFile = %q, want %q", got.ConfigFile, file)
	}

	// 该 viper 使用自定义 KeyDelimiter（避免 apollo 命名空间的 / 被拆分），
	// 读取断言走 AllSettings 的嵌套结构，不按 . 取 key。
	settings := got.Viper.AllSettings()
	frame, ok := settings["frame"].(map[string]interface{})
	if !ok {
		t.Fatalf("frame 分片缺失: %#v", settings)
	}
	if debug, ok := frame["debug"]; !ok || debug != false {
		t.Errorf("frame.debug = %#v, want false", frame["debug"])
	}
	myconf, ok := settings["myconf"].(map[string]interface{})
	if !ok || myconf["a"] != 1 {
		t.Errorf("myconf = %#v, want {a:1}", settings["myconf"])
	}
	if got.NeedApolloProvider {
		t.Error("无 apollo 分片时不应启用 apollo_provider")
	}
}

func TestResolveApolloSectionInDefaultFile(t *testing.T) {
	dir := t.TempDir()
	file := writeConfigFile(t, dir, "default.yaml", "apollo:\n  Address: http://apollo:8080\n")

	got, err := Resolve(Source{AppName: "svc", DefaultFiles: []string{file}})
	if err != nil {
		t.Fatalf("Resolve 失败: %v", err)
	}
	if !got.NeedApolloProvider {
		t.Error("配置含 apollo 分片时应启用 apollo_provider")
	}
}

func TestResolveDisableApolloFallsBackToDefaultFile(t *testing.T) {
	dir := t.TempDir()
	file := writeConfigFile(t, dir, "default.yaml", "myconf:\n  A: 2\n")

	got, err := Resolve(Source{
		AppName: "svc",
		Apollo: &ApolloConfig{
			Address:           "http://unreachable",
			DisableUAppApollo: true,
			DisableAppApollo:  true,
		},
		DefaultFiles: []string{file},
	})
	if err != nil {
		t.Fatalf("两侧 apollo 都关闭时不应连接 apollo: %v", err)
	}
	if got.ConfigFile != file {
		t.Errorf("ConfigFile = %q, want %q", got.ConfigFile, file)
	}
	myconf, ok := got.Viper.AllSettings()["myconf"].(map[string]interface{})
	if !ok || myconf["a"] != 2 {
		t.Errorf("myconf = %#v, want {a:2}", got.Viper.AllSettings()["myconf"])
	}
	if got.NeedApolloProvider {
		t.Error("不应启用 apollo_provider")
	}
}

func TestResolveMissingDefaultFiles(t *testing.T) {
	got, err := Resolve(Source{
		AppName:      "svc",
		DefaultFiles: []string{filepath.Join(t.TempDir(), "nope.yaml")},
	})
	if err != nil {
		t.Fatalf("Resolve 失败: %v", err)
	}
	if got.ConfigFile != "" {
		t.Errorf("ConfigFile 应为空, got %q", got.ConfigFile)
	}
	if got.NeedApolloProvider {
		t.Error("不应启用 apollo_provider")
	}
}

func TestIsDaemonCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"无参数", nil, false},
		{"只有程序名", []string{"app"}, false},
		{"普通参数", []string{"app", "run"}, false},
		{"install", []string{"app", "install"}, true},
		{"-start", []string{"app", "-start"}, true},
		{"status", []string{"app", "status"}, true},
		{"uninstall", []string{"app", "uninstall"}, true},
		{"-h 不是守护命令", []string{"app", "-h"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDaemonCommand(tt.args); got != tt.want {
				t.Errorf("IsDaemonCommand(%#v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
