package app

import "testing"

// TestNew 验证组合根能完成一次真实装配。
// zapp 会在进程级注册配置 flag，因此本测试只装配一次，不做并发。
func TestNew(t *testing.T) {
	t.Setenv("ApolloAddress", "")

	app := New("uapp-test",
		WithoutDaemon(),
		WithoutDefaultPlugins(),
		WithoutHTTPClientReplace(),
	)
	if app == nil {
		t.Fatal("New 返回 nil")
	}
	if app.Name() != "uapp-test" {
		t.Errorf("Name = %q, want uapp-test", app.Name())
	}
}
