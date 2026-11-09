package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	zconfig "github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
)

// keyDelimiter 让 key 中的 / 不被当作层级分隔符（apollo 命名空间形如 frame.json）。
const keyDelimiter = `\/empty_delimiter\/`

// Source 是一次配置解析的来源声明。
type Source struct {
	AppName      string        // 应用名，用于缺省 AppId 与日志文件名
	Apollo       *ApolloConfig // nil 表示不使用 Apollo
	DefaultFiles []string      // 默认配置文件候选列表，按顺序取第一个存在的
}

// Resolved 是配置解析结果。
type Resolved struct {
	Viper              *viper.Viper // 交给 zapp 的配置源
	NeedApolloProvider bool         // 是否需要启用 apollo_provider 插件
	ConfigFile         string       // 实际加载的默认配置文件，空表示未加载
}

// Resolve 按「uapp 配置 → 应用配置」的顺序合并配置来源。
//
// 规则：
//  1. 启用 uapp 侧 Apollo 时，uapp 配置取自 Apollo；否则取自默认配置文件。
//  2. 启用应用侧 Apollo 时，把 Apollo 参数写进 vi 的 apollo 分片交给 zapp 加载；
//     否则应用配置取自默认配置文件并覆盖 uapp 配置。
//  3. 最终配置里出现 apollo 分片，就需要 apollo_provider 插件。
func Resolve(s Source) (*Resolved, error) {
	vi := newViper()

	if s.Apollo != nil && !s.Apollo.DisableUAppApollo {
		uAppConf := zconfig.NewConfig(s.AppName, zconfig.WithApollo(s.Apollo.uappZappConfig()), zconfig.WithoutFlag())
		settings := uAppConf.GetViper().AllSettings()
		// 去掉 apollo 分片，否则 zapp 启动时会再拉一次 uapp 配置。
		delete(settings, consts.ApolloConfigKey)
		if err := vi.MergeConfigMap(settings); err != nil {
			return nil, fmt.Errorf("合并 uapp 配置失败: %w", err)
		}
	}

	if s.Apollo != nil && !s.Apollo.DisableAppApollo {
		vi.Set(consts.ApolloConfigKey, s.Apollo.appZappConfig(s.AppName))
		return &Resolved{Viper: vi, NeedApolloProvider: true}, nil
	}

	fileVi, file, err := loadFirstExisting(s.DefaultFiles)
	if err != nil {
		return nil, err
	}
	if file != "" {
		if err := vi.MergeConfigMap(fileVi.AllSettings()); err != nil {
			return nil, fmt.Errorf("合并默认配置文件失败: %w", err)
		}
	}

	return &Resolved{
		Viper:              vi,
		NeedApolloProvider: vi.IsSet(consts.ApolloConfigKey),
		ConfigFile:         file,
	}, nil
}

// IsDaemonCommand 判断命令行首个参数是否为守护进程命令。
//
// 这类命令由 zapp 的 daemon 模型在配置加载之前处理并退出，
// uapp 必须跳过自己的配置解析，否则配置缺失时 install/stop 无法执行。
func IsDaemonCommand(args []string) bool {
	if len(args) < 2 {
		return false
	}
	_, ok := daemonCommands[args[1]]
	return ok
}

var daemonCommands = map[string]struct{}{
	"install": {}, "remove": {}, "start": {}, "stop": {}, "restart": {}, "status": {}, "uninstall": {},
	"-install": {}, "-remove": {}, "-start": {}, "-stop": {}, "-restart": {}, "-status": {}, "-uninstall": {},
}

func newViper() *viper.Viper {
	return viper.NewWithOptions(viper.KeyDelimiter(keyDelimiter))
}

// loadFirstExisting 按顺序取第一个存在的配置文件并加载。
func loadFirstExisting(files []string) (*viper.Viper, string, error) {
	vi := newViper()
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		_, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, "", fmt.Errorf("读取配置文件信息失败 %s: %w", file, err)
		}

		vi.SetConfigFile(file)
		if err = vi.MergeInConfig(); err != nil {
			return nil, "", fmt.Errorf("合并配置文件失败 %s: %w", file, err)
		}
		return vi, file, nil
	}
	return vi, "", nil
}
