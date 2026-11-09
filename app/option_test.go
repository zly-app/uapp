package app

import (
	"testing"

	"github.com/zly-app/zapp"

	uconfig "github.com/zly-app/uapp/model/config"
)

func TestNewOptionsDefaults(t *testing.T) {
	t.Setenv("ApolloAddress", "")

	o := newOptions()
	if !o.defaultPlugins {
		t.Error("默认应启用默认插件")
	}
	if !o.daemon {
		t.Error("默认应启用守护进程")
	}
	if !o.replaceHTTPClient {
		t.Error("默认应替换 http 标准库客户端")
	}
	if o.apollo != nil {
		t.Error("未设置 ApolloAddress 时 apollo 应为 nil")
	}
	if len(o.configFiles) != 4 {
		t.Errorf("默认配置文件候选数 = %d, want 4", len(o.configFiles))
	}
	if len(o.zappOpts) != 0 {
		t.Errorf("默认不应有额外 zapp 选项, got %d", len(o.zappOpts))
	}
}

func TestOptions(t *testing.T) {
	t.Setenv("ApolloAddress", "")

	cfg := &uconfig.ApolloConfig{Address: "http://apollo:8080"}
	o := newOptions(
		WithoutDefaultPlugins(),
		WithoutDaemon(),
		WithoutHTTPClientReplace(),
		WithApolloConfig(cfg),
		WithConfigFiles("a.yaml", "b.yaml"),
		WithZappOptions(zapp.WithEnableDaemon(), zapp.WithIgnoreInjectOfDisablePlugin()),
	)

	if o.defaultPlugins || o.daemon || o.replaceHTTPClient {
		t.Error("三个开关应全部关闭")
	}
	if o.apollo != cfg {
		t.Error("显式 Apollo 配置未生效")
	}
	if len(o.configFiles) != 2 || o.configFiles[0] != "a.yaml" {
		t.Errorf("configFiles = %#v", o.configFiles)
	}
	if len(o.zappOpts) != 2 {
		t.Errorf("zappOpts 数量 = %d, want 2", len(o.zappOpts))
	}
}

func TestNewOptionsApolloFromEnv(t *testing.T) {
	t.Setenv("ApolloAddress", "http://apollo:8080")
	t.Setenv("ApolloUAppID", "uapp-x")

	o := newOptions()
	if o.apollo == nil || o.apollo.Address != "http://apollo:8080" || o.apollo.UAppID != "uapp-x" {
		t.Errorf("环境变量未正确进入默认选项: %#v", o.apollo)
	}
}
