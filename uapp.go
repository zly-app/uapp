package uapp

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/zly-app/component/http"
	"github.com/zly-app/plugin/otlp"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/plugin/apollo_provider"
	"go.uber.org/zap"

	"github.com/zly-app/plugin/pprof"
)

func NewApp(appName string, opts ...zapp.Option) core.IApp {
	if appName == "" {
		log.Fatal("appName is empty")
	}

	initConfig()

	allOpts := []zapp.Option{
		zapp.WithEnableDaemon(),                     // 启用守护进程
		zapp.WithIgnoreInjectOfDisablePlugin(true),  // 忽略未启用的插件注入
		zapp.WithIgnoreInjectOfDisableService(true), // 忽略未启用的服务注入

		pprof.WithPlugin(), // pprof
		otlp.WithPlugin(),
	}

	// 兼容 WithEnableDaemon 参数
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "install", "remove", "start", "stop", "restart", "status", "uninstall",
			"-install", "-remove", "-start", "-stop", "-restart", "-status", "-uninstall", "-h":
			return zapp.NewApp(appName, allOpts...)
		}
	}

	uAppOpts := makeUAppOpts(appName)
	allOpts = append(allOpts, uAppOpts...)

	allOpts = append(allOpts, opts...)
	app := zapp.NewApp(appName, allOpts...)

	http.ReplaceStd()

	return app
}

func NewAppNotPlugins(appName string, opts ...zapp.Option) core.IApp {
	if appName == "" {
		log.Fatal("appName is empty")
	}

	initConfig()

	allOpts := []zapp.Option{}

	// 兼容 WithEnableDaemon 参数
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "install", "remove", "start", "stop", "restart", "status", "uninstall",
			"-install", "-remove", "-start", "-stop", "-restart", "-status", "-uninstall", "-h":
			return zapp.NewApp(appName, allOpts...)
		}
	}

	uAppOpts := makeUAppOpts(appName)
	allOpts = append(allOpts, uAppOpts...)

	allOpts = append(allOpts, opts...)
	app := zapp.NewApp(appName, allOpts...)

	http.ReplaceStd()

	return app
}

// 生成uApp选项
func makeUAppOpts(appName string) []zapp.Option {
	cl := getConfig()
	vi := newViper()

	allowApollo := *cl.ApolloAddress != ""

	// uapp 配置
	if allowApollo && !*cl.ApolloDisableApolloUApp {
		// uapp配置
		uAppApolloConfig := &config.ApolloConfig{
			Address:                 *cl.ApolloAddress,
			AppId:                   utils.Ternary.Or(*cl.ApolloUAppID, "uapp").(string),
			AccessKey:               *cl.ApolloAccessKey,
			AuthBasicUser:           *cl.ApolloAuthBasicUser,
			AuthBasicPassword:       *cl.ApolloAuthBasicPassword,
			Cluster:                 utils.Ternary.Or(*cl.ApolloCluster, "default").(string),
			AlwaysLoadFromRemote:    *cl.ApolloAlwaysLoadFromRemote,
			BackupFile:              *cl.ApolloBackupFile,
			ApplicationDataType:     *cl.ApolloApplicationDataType,
			ApplicationParseKeys:    nil,
			Namespaces:              nil,
			IgnoreNamespaceNotFound: false,
		}
		if uAppApolloConfig.BackupFile != "" {
			uAppApolloConfig.BackupFile += ".uapp"
		}

		uAppConf := config.NewConfig(appName, config.WithApollo(uAppApolloConfig), config.WithoutFlag())
		uAppConfigs := uAppConf.GetViper().AllSettings()
		// 这里要去掉 apollo 配置, 否则zapp启动时仍然会从apollo中获取一次
		delete(uAppConfigs, consts.ApolloConfigKey)
		// 合并uapp配置
		err := vi.MergeConfigMap(uAppConfigs)
		if err != nil {
			log.Log.Fatal("合并'uapp配置'时错误", zap.Error(err))
		}
	}

	// 应用配置, 这里仍然允许用户通过命令行覆盖配置
	if allowApollo && !*cl.ApolloDisableApolloApp {
		appApolloConfig := &config.ApolloConfig{
			Address:                 *cl.ApolloAddress,
			AppId:                   utils.Ternary.Or(*cl.ApolloAppId, appName).(string),
			AccessKey:               *cl.ApolloAccessKey,
			AuthBasicUser:           *cl.ApolloAuthBasicUser,
			AuthBasicPassword:       *cl.ApolloAuthBasicPassword,
			Cluster:                 utils.Ternary.Or(*cl.ApolloCluster, "default").(string),
			AlwaysLoadFromRemote:    *cl.ApolloAlwaysLoadFromRemote,
			BackupFile:              *cl.ApolloBackupFile,
			ApplicationDataType:     *cl.ApolloApplicationDataType,
			IgnoreNamespaceNotFound: *cl.ApolloIgnoreNamespaceNotFound,
		}
		applicationParseKeys := *cl.ApolloApplicationParseKeys
		if applicationParseKeys != "" {
			appApolloConfig.ApplicationParseKeys = strings.Split(applicationParseKeys, ",")
		}
		namespaces := *cl.ApolloNamespaces
		if namespaces != "" {
			appApolloConfig.Namespaces = strings.Split(namespaces, ",")
		}

		// 这里不立即获取配置数据, 而是包装为opts交给使用者, 因为用户可能通过命令行-conf覆盖配置
		vi.Set(consts.ApolloConfigKey, appApolloConfig) // 为应用配置写入apollo依据
		opts := []zapp.Option{
			zapp.WithConfigOption(config.WithViper(vi)), // 告诉zapp从这个vi中加载配置
			apollo_provider.WithPlugin(true),            // 配置观察提供者--apollo
		}
		return opts
	}

	// 当没有启用应用配置时, 读取默认配置文件的数据并覆盖uapp, 允许用户通过命令行-conf覆盖所有配置
	appConf := loadDefaultFiles()
	err := vi.MergeConfigMap(appConf.AllSettings())
	if err != nil {
		log.Log.Fatal("合并用户默认配置文件数据时错误", zap.Error(err))
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

// 加载默认配置文件, 默认配置文件不存在返回nil
func loadDefaultFiles() *viper.Viper {
	files := strings.Split(consts.DefaultConfigFiles, ",")
	vi := newViper()
	for _, file := range files {
		_, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Log.Fatal("读取配置文件信息失败", zap.String("file", file), zap.Error(err))
		}

		vi.SetConfigFile(file)
		if err = vi.MergeInConfig(); err != nil {
			log.Log.Fatal("合并配置文件失败", zap.String("file", file), zap.Error(err))
		}
		log.Log.Info("使用默认配置文件", zap.String("file", file))
		return vi
	}
	return vi
}

func newViper() *viper.Viper {
	return viper.NewWithOptions(viper.KeyDelimiter(`\/empty_delimiter\/`))
}
