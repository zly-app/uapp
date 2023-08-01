package uapp

import (
	"os"
	"strings"

	"github.com/spf13/cast"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

func NewApp(appName string, opts ...zapp.Option) core.IApp {
	apolloConfig := &config.ApolloConfig{
		Address:                 os.Getenv("ApolloAddress"),
		AppId:                   utils.Ternary.Or(os.Getenv("ApolloAppId"), appName).(string),
		AccessKey:               os.Getenv("ApolloAccessKey"),
		AuthBasicUser:           os.Getenv("ApolloAuthBasicUser"),
		AuthBasicPassword:       os.Getenv("ApolloAuthBasicPassword"),
		Cluster:                 utils.Ternary.Or(os.Getenv("ApolloCluster"), "default").(string),
		AlwaysLoadFromRemote:    cast.ToBool(os.Getenv("ApolloAlwaysLoadFromRemote")),
		BackupFile:              os.Getenv("ApolloBackupFile"),
		ApplicationDataType:     os.Getenv("ApolloApplicationDataType"),
		IgnoreNamespaceNotFound: cast.ToBool(os.Getenv("ApolloIgnoreNamespaceNotFound")),
	}
	applicationParseKeys := os.Getenv("ApolloApplicationParseKeys")
	if applicationParseKeys != "" {
		apolloConfig.ApplicationParseKeys = strings.Split(applicationParseKeys, ",")
	}
	namespaces := os.Getenv("ApolloNamespaces")
	if namespaces != "" {
		apolloConfig.Namespaces = strings.Split(namespaces, ",")
	}

	allOpts := []zapp.Option{
		zapp.WithCustomComponent(makeCustomComponent), // 自定义组件
		zapp.WithEnableDaemon(),                       // 启用守护进程
		zapp.WithIgnoreInjectOfDisablePlugin(true),    // 忽略未启用的插件注入
		zapp.WithIgnoreInjectOfDisableService(true),   // 忽略未启用的服务注入

		zapp.WithConfigOption(config.WithApollo(apolloConfig)), // 添加apollo配置

		// zipokinotel
		// honey
	}
	allOpts = append(allOpts, opts...)
	app := zapp.NewApp(appName, allOpts...)
	return app
}
