package uapp

import (
	"os"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/zly-app/plugin/honey"
	"github.com/zly-app/plugin/zipkinotel"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/plugin/apollo_provider"
	"go.uber.org/zap"
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

	uAppOpts := makeUAppOpts(appName)
	allOpts = append(allOpts, uAppOpts...)

	allOpts = append(allOpts, opts...)
	app := zapp.NewApp(appName, allOpts...)
	return app
}

// 生成uApp选项
func makeUAppOpts(appName string) []zapp.Option {
	vi := viper.New()

	backupFile := os.Getenv("ApolloBackupFile")
	if backupFile != "" {
		backupFile += ".uapp"
	}

	// uapp配置
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
	if uAppApolloConfig.Address != "" {
		uAppConf := config.NewConfig(uAppApolloConfig.AppId, config.WithApollo(uAppApolloConfig), config.WithoutFlag(uAppApolloConfig))
		uAppConfigs := uAppConf.GetViper().AllSettings()
		// 这里要去掉 apollo 配置, 否则zapp启动时仍然会从apollo中获取一次
		delete(uAppConfigs, consts.ApolloConfigKey)
		// 合并uapp配置
		err := vi.MergeConfigMap(uAppConfigs)
		if err != nil {
			logger.Log.Fatal("合并'uapp配置'时错误", zap.Error(err))
		}
	}

	// 应用配置
	zAppApolloConfig, ok := getApolloConfigFromEnv(appName)
	if ok {
		vi.Set(consts.ApolloConfigKey, zAppApolloConfig) // 为应用配置写入apollo依据
		opts := []zapp.Option{
			zapp.WithConfigOption(config.WithViper(vi)), // 告诉zapp从这个vi中加载配置
			apollo_provider.WithPlugin(true),            // 配置观察提供者--apollo
		}
		return opts
	}

	// 用户自行处理应用配置
	appConf := config.NewConfig(uAppApolloConfig.AppId, config.WithoutFlag(uAppApolloConfig))
	err := vi.MergeConfigMap(appConf.GetViper().AllSettings())
	if err != nil {
		logger.Log.Fatal("合并用户自行处理的'应用配置'时错误", zap.Error(err))
	}

	opts := []zapp.Option{
		zapp.WithConfigOption(config.WithViper(vi)), // 写入apollo中获取的配置
	}
	// 如果有 apollo 配置则开启 "配置观察提供者--apollo" 插件
	if vi.IsSet(consts.ApolloConfigKey) {
		opts = append(opts,
			apollo_provider.WithPlugin(true), // 配置观察提供者--apollo
		)
	}
	return opts
}

// 从环境变量中获取apollo配置
func getApolloConfigFromEnv(appName string) (*config.ApolloConfig, bool) {
	disable := os.Getenv("ApolloDisableApolloApp")
	if strings.ToLower(disable) == "true" {
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
		IgnoreNamespaceNotFound: cast.ToBool(os.Getenv("ApolloIgnoreNamespaceNotFound")),
	}
	if apolloConfig.Address == "" {
		return nil, false
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
