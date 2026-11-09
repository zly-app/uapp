// Package uapp 是 zapp 的启动器：统一入口 New，预置常用配置来源与默认插件。
//
// 本包是根门面，只做类型与函数转发；装配在 app，配置规则在 model/config。
package uapp

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"

	"github.com/zly-app/uapp/app"
	uconfig "github.com/zly-app/uapp/model/config"
)

type (
	// Option 是装配期选项。
	Option = app.Option
	// ApolloConfig 是 Apollo 连接参数。
	ApolloConfig = uconfig.ApolloConfig
)

// New 创建应用。
func New(appName string, opts ...Option) core.IApp {
	return app.New(appName, opts...)
}

// WithoutDefaultPlugins 不启用默认插件（pprof / otlp）。
func WithoutDefaultPlugins() Option {
	return app.WithoutDefaultPlugins()
}

// WithoutDaemon 不启用守护进程。
func WithoutDaemon() Option {
	return app.WithoutDaemon()
}

// WithoutHTTPClientReplace 不替换 http.DefaultClient 与 http.DefaultTransport。
func WithoutHTTPClientReplace() Option {
	return app.WithoutHTTPClientReplace()
}

// WithApolloConfig 显式指定 Apollo 参数，覆盖环境变量。
func WithApolloConfig(conf *ApolloConfig) Option {
	return app.WithApolloConfig(conf)
}

// WithConfigFiles 覆盖默认配置文件候选列表。
func WithConfigFiles(files ...string) Option {
	return app.WithConfigFiles(files...)
}

// WithZappOptions 追加 zapp 选项（例如 grpc 的 g.Options()）。
func WithZappOptions(opts ...zapp.Option) Option {
	return app.WithZappOptions(opts...)
}
