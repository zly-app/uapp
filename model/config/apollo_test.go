package config

import (
	"reflect"
	"testing"
)

func TestLoadApolloFromEnvDisabled(t *testing.T) {
	t.Setenv("ApolloAddress", "")
	if got := LoadApolloFromEnv(); got != nil {
		t.Fatalf("未设置 ApolloAddress 时应返回 nil, got %#v", got)
	}
}

func TestLoadApolloFromEnv(t *testing.T) {
	t.Setenv("ApolloAddress", "http://apollo:8080")
	t.Setenv("ApolloUAppID", "")
	t.Setenv("ApolloAppId", "")
	t.Setenv("ApolloAccessKey", "key")
	t.Setenv("ApolloCluster", "")
	t.Setenv("ApolloBackupFile", "./backup.apollo")
	t.Setenv("ApolloApplicationDataType", "")
	t.Setenv("ApolloApplicationParseKeys", "a, b")
	t.Setenv("ApolloNamespaces", "ns1,ns2")
	t.Setenv("ApolloDisableApolloUApp", "true")
	t.Setenv("ApolloDisableApolloApp", "false")
	t.Setenv("ApolloIgnoreNamespaceNotFound", "true")

	got := LoadApolloFromEnv()
	if got == nil {
		t.Fatal("应返回配置")
	}
	if got.Address != "http://apollo:8080" {
		t.Errorf("Address = %q", got.Address)
	}
	if got.UAppID != defaultUAppID {
		t.Errorf("UAppID 默认值 = %q, want %q", got.UAppID, defaultUAppID)
	}
	if got.AppID != "" {
		t.Errorf("AppID 应留空由 appName 兜底, got %q", got.AppID)
	}
	if got.Cluster != defaultCluster {
		t.Errorf("Cluster 默认值 = %q, want %q", got.Cluster, defaultCluster)
	}
	if got.ApplicationDataType != defaultApplicationDataType {
		t.Errorf("ApplicationDataType 默认值 = %q", got.ApplicationDataType)
	}
	if !reflect.DeepEqual(got.ApplicationParseKeys, []string{"a", "b"}) {
		t.Errorf("ApplicationParseKeys = %#v", got.ApplicationParseKeys)
	}
	if !reflect.DeepEqual(got.Namespaces, []string{"ns1", "ns2"}) {
		t.Errorf("Namespaces = %#v", got.Namespaces)
	}
	if !got.DisableUAppApollo || got.DisableAppApollo {
		t.Errorf("Disable 开关解析错误: UApp=%v App=%v", got.DisableUAppApollo, got.DisableAppApollo)
	}
	if !got.IgnoreNamespaceNotFound {
		t.Error("IgnoreNamespaceNotFound 应为 true")
	}
}

func TestUAppZappConfig(t *testing.T) {
	c := &ApolloConfig{
		Address:                 "addr",
		BackupFile:              "./backup.apollo",
		IgnoreNamespaceNotFound: true,
	}

	got := c.uappZappConfig()
	if got.AppId != defaultUAppID {
		t.Errorf("AppId = %q, want %q", got.AppId, defaultUAppID)
	}
	if got.BackupFile != "./backup.apollo"+uappBackupFileSuffix {
		t.Errorf("BackupFile = %q", got.BackupFile)
	}
	if got.Cluster != defaultCluster {
		t.Errorf("Cluster = %q", got.Cluster)
	}
	if got.ApplicationDataType != defaultApplicationDataType {
		t.Errorf("ApplicationDataType = %q", got.ApplicationDataType)
	}
	if got.Namespaces != nil || got.ApplicationParseKeys != nil {
		t.Error("uapp 侧不应携带额外命名空间/解析 key")
	}
	if got.IgnoreNamespaceNotFound {
		t.Error("uapp 侧应固定忽略开关为 false")
	}
}

func TestAppZappConfig(t *testing.T) {
	c := &ApolloConfig{
		Address:                 "addr",
		AppID:                   "",
		BackupFile:              "./backup.apollo",
		ApplicationParseKeys:    []string{"k"},
		Namespaces:              []string{"ns"},
		IgnoreNamespaceNotFound: true,
	}

	got := c.appZappConfig("my-service")
	if got.AppId != "my-service" {
		t.Errorf("AppId = %q, want my-service", got.AppId)
	}
	if got.BackupFile != "./backup.apollo" {
		t.Errorf("应用侧 BackupFile 不应加后缀, got %q", got.BackupFile)
	}
	if !reflect.DeepEqual(got.ApplicationParseKeys, []string{"k"}) {
		t.Errorf("ApplicationParseKeys = %#v", got.ApplicationParseKeys)
	}
	if !reflect.DeepEqual(got.Namespaces, []string{"ns"}) {
		t.Errorf("Namespaces = %#v", got.Namespaces)
	}
	if !got.IgnoreNamespaceNotFound {
		t.Error("应用侧应保留 IgnoreNamespaceNotFound")
	}

	if got := (&ApolloConfig{AppID: "fixed"}).appZappConfig("my-service"); got.AppId != "fixed" {
		t.Errorf("显式 AppID 不应被 appName 覆盖, got %q", got.AppId)
	}
}
