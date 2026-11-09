// Package app 是 uapp 的组合根：唯一装配点。
//
// 它把「配置解析（model/config）」「默认插件（pprof / otlp / apollo_provider）」
// 「zapp 应用」和「http 标准库替换」装配到一起；不含配置合并规则。
package app

import (
	"os"

	"github.com/zly-app/component/http"
	"github.com/zly-app/plugin/otlp"
	"github.com/zly-app/plugin/pprof"
	"github.com/zly-app/zapp"
	zconfig "github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/plugin/apollo_provider"
	"go.uber.org/zap"

	uconfig "github.com/zly-app/uapp/model/config"
)

// New 创建应用。配置解析失败直接 Fatal（启动期失败没有可恢复路径）。
func New(appName string, opts ...Option) core.IApp {
	if appName == "" {
		log.Fatal("appName is empty")
	}

	o := newOptions(opts...)
	zappOpts := make([]zapp.Option, 0, 12)

	if o.daemon {
		zappOpts = append(zappOpts, zapp.WithEnableDaemon())
	}
	zappOpts = append(zappOpts,
		zapp.WithIgnoreInjectOfDisablePlugin(true),
		zapp.WithIgnoreInjectOfDisableService(true),
	)

	if o.defaultPlugins {
		zappOpts = append(zappOpts, pprof.WithPlugin(), otlp.WithPlugin())
	}

	zappOpts = append(zappOpts, buildConfigOptions(appName, o)...)
	zappOpts = append(zappOpts, o.zappOpts...)

	app := zapp.NewApp(appName, zappOpts...)

	if o.replaceHTTPClient {
		http.ReplaceStd()
	}
	return app
}

// buildConfigOptions 把配置解析结果转成 zapp 选项。
func buildConfigOptions(appName string, o *options) []zapp.Option {
	// 守护命令由 zapp 的 daemon 模型在配置加载前处理并退出，
	// 这里跳过 uapp 配置解析，保证配置缺失时 install/stop 仍可执行。
	if uconfig.IsDaemonCommand(os.Args) {
		return nil
	}

	resolved, err := uconfig.Resolve(uconfig.Source{
		AppName:      appName,
		Apollo:       o.apollo,
		DefaultFiles: o.configFiles,
	})
	if err != nil {
		log.Fatal("加载 uapp 配置失败", zap.Error(err))
	}
	if resolved.ConfigFile != "" {
		log.Log.Info("使用默认配置文件", zap.String("file", resolved.ConfigFile))
	}

	opts := []zapp.Option{zapp.WithConfigOption(zconfig.WithViper(resolved.Viper))}
	if resolved.NeedApolloProvider {
		opts = append(opts, apollo_provider.WithPlugin(true))
	}
	return opts
}
