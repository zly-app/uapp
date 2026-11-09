package app

import (
	"strings"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/consts"

	uconfig "github.com/zly-app/uapp/model/config"
)

// Option 是 uapp 的装配期选项。不传选项时使用默认装配（默认插件、守护进程、http 替换全开）。
type Option func(o *options)

type options struct {
	defaultPlugins    bool
	daemon            bool
	replaceHTTPClient bool
	apollo            *uconfig.ApolloConfig
	configFiles       []string
	zappOpts          []zapp.Option
}

func newOptions(opts ...Option) *options {
	o := &options{
		defaultPlugins:    true,
		daemon:            true,
		replaceHTTPClient: true,
		apollo:            uconfig.LoadApolloFromEnv(),
		configFiles:       strings.Split(consts.DefaultConfigFiles, ","),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithoutDefaultPlugins 不启用默认插件（pprof / otlp）。
func WithoutDefaultPlugins() Option {
	return func(o *options) {
		o.defaultPlugins = false
	}
}

// WithoutDaemon 不启用守护进程。
func WithoutDaemon() Option {
	return func(o *options) {
		o.daemon = false
	}
}

// WithoutHTTPClientReplace 不替换 http.DefaultClient 与 http.DefaultTransport。
func WithoutHTTPClientReplace() Option {
	return func(o *options) {
		o.replaceHTTPClient = false
	}
}

// WithApolloConfig 显式指定 Apollo 参数，覆盖环境变量。
func WithApolloConfig(conf *uconfig.ApolloConfig) Option {
	return func(o *options) {
		o.apollo = conf
	}
}

// WithConfigFiles 覆盖默认配置文件候选列表。
func WithConfigFiles(files ...string) Option {
	return func(o *options) {
		o.configFiles = files
	}
}

// WithZappOptions 追加 zapp 选项（例如 grpc 的 g.Options()）。
func WithZappOptions(opts ...zapp.Option) Option {
	return func(o *options) {
		o.zappOpts = append(o.zappOpts, opts...)
	}
}
