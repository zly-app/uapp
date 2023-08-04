package uapp

import (
	"os"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/zly-app/plugin/honey"
	"github.com/zly-app/plugin/zipkinotel"
	"github.com/zlyuancn/zstr"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

func NewApp(appName string, opts ...zapp.Option) core.IApp {
	allOpts := []zapp.Option{
		zapp.WithCustomComponent(makeCustomComponent), // 自定义组件
		zapp.WithEnableDaemon(),                       // 启用守护进程
		zapp.WithIgnoreInjectOfDisablePlugin(true),    // 忽略未启用的插件注入
		zapp.WithIgnoreInjectOfDisableService(true),   // 忽略未启用的服务注入

		zipkinotel.WithPlugin(), // trace
		honey.WithPlugin(),      // log
	}

	// 添加apollo配置
	if confVi, ok := makeUAppConfig(appName); ok {
		allOpts = append(allOpts, zapp.WithConfigOption(config.WithViper(confVi)))
	}

	allOpts = append(allOpts, opts...)
	app := zapp.NewApp(appName, allOpts...)
	return app
}

// 生成uApp配置
func makeUAppConfig(appName string) (*viper.Viper, bool) {
	backupFile := os.Getenv("ApolloBackupFile")
	if backupFile != "" {
		backupFile += ".uapp"
	}
	uAppApolloConfig := &config.ApolloConfig{
		Address:                 os.Getenv("ApolloAddress"),
		AppId:                   utils.Ternary.Or(os.Getenv("ApolloUAppID"), "uapp").(string),
		AccessKey:               os.Getenv("ApolloAccessKey"),
		AuthBasicUser:           os.Getenv("ApolloAuthBasicUser"),
		AuthBasicPassword:       os.Getenv("ApolloAuthBasicPassword"),
		Cluster:                 utils.Ternary.Or(os.Getenv("ApolloCluster"), "default").(string),
		AlwaysLoadFromRemote:    cast.ToBool(os.Getenv("ApolloAlwaysLoadFromRemote")),
		BackupFile:              backupFile,
		ApplicationDataType:     os.Getenv("ApolloApplicationDataType"),
		ApplicationParseKeys:    nil,
		Namespaces:              nil,
		IgnoreNamespaceNotFound: false,
	}
	if uAppApolloConfig.Address == "" {
		return nil, false
	}

	conf := config.NewConfig(uAppApolloConfig.AppId, config.WithApollo(uAppApolloConfig), config.WithoutFlag(uAppApolloConfig))
	vi := conf.GetViper()

	zAppApolloConfig, ok := getApolloConfigFromEnv(appName)
	if ok {
		vi.Set(consts.ApolloConfigKey, zAppApolloConfig)
	}
	return vi, true
}

// 从环境变量中获取apollo配置
func getApolloConfigFromEnv(appName string) (*config.ApolloConfig, bool) {
	ok := zstr.GetBool(os.Getenv("ApolloAppDisable"))
	if !ok {
		return nil, false
	}

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
		IgnoreNamespaceNotFound: true,
	}
	applicationParseKeys := os.Getenv("ApolloApplicationParseKeys")
	if applicationParseKeys != "" {
		apolloConfig.ApplicationParseKeys = strings.Split(applicationParseKeys, ",")
	}
	namespaces := os.Getenv("ApolloNamespaces")
	if namespaces != "" {
		apolloConfig.Namespaces = strings.Split(namespaces, ",")
	}
	return apolloConfig, true
}
