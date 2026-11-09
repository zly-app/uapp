// Package config 是 uapp 的配置模型：独占 uapp 的配置数据，以及
// 「uapp 配置 → 应用配置」的合并规则。
//
// 允许调用：foundation/env（取值机制）、zapp/config（配置加载机制）。
// 禁止：import app 或根包；禁止持有进程级可变状态。
package config

import (
	"github.com/zly-app/uapp/foundation/env"
	zconfig "github.com/zly-app/zapp/config"
)

const (
	defaultUAppID              = "uapp"
	defaultCluster             = "default"
	defaultApplicationDataType = "yaml"
	uappBackupFileSuffix       = ".uapp"
)

// ApolloConfig 是 uapp 连接 Apollo 的参数。零值表示不启用 Apollo。
type ApolloConfig struct {
	Address string // apollo-api 地址，多个地址用英文逗号连接
	UAppID  string // uapp 应用名，默认 uapp
	AppID   string // 当前应用名，默认取 New 传入的 appName

	AccessKey         string // 验证 key，优先级高于基础认证
	AuthBasicUser     string // 基础认证用户名
	AuthBasicPassword string // 基础认证密码
	Cluster           string // 集群名，默认 default

	AlwaysLoadFromRemote    bool     // 总是从远程获取，远程失败时不回退备份文件
	BackupFile              string   // 备份文件名；uapp 侧自动加 .uapp 后缀
	ApplicationDataType     string   // application 命名空间下 key 的数据类型，默认 yaml
	ApplicationParseKeys    []string // application 命名空间下额外解析的 key
	Namespaces              []string // 其他自定义命名空间
	IgnoreNamespaceNotFound bool     // 忽略其他自定义命名空间不存在

	DisableUAppApollo bool // uapp 不从 Apollo 取 uapp 配置
	DisableAppApollo  bool // uapp 不从 Apollo 取应用配置
}

// LoadApolloFromEnv 从环境变量构造 ApolloConfig；未设置 Address 时返回 nil。
func LoadApolloFromEnv() *ApolloConfig {
	address := env.String("ApolloAddress", "")
	if address == "" {
		return nil
	}

	return &ApolloConfig{
		Address:                 address,
		UAppID:                  env.String("ApolloUAppID", defaultUAppID),
		AppID:                   env.String("ApolloAppId", ""),
		AccessKey:               env.String("ApolloAccessKey", ""),
		AuthBasicUser:           env.String("ApolloAuthBasicUser", ""),
		AuthBasicPassword:       env.String("ApolloAuthBasicPassword", ""),
		Cluster:                 env.String("ApolloCluster", defaultCluster),
		AlwaysLoadFromRemote:    env.Bool("ApolloAlwaysLoadFromRemote", false),
		BackupFile:              env.String("ApolloBackupFile", ""),
		ApplicationDataType:     env.String("ApolloApplicationDataType", defaultApplicationDataType),
		ApplicationParseKeys:    env.List("ApolloApplicationParseKeys"),
		Namespaces:              env.List("ApolloNamespaces"),
		IgnoreNamespaceNotFound: env.Bool("ApolloIgnoreNamespaceNotFound", false),
		DisableUAppApollo:       env.Bool("ApolloDisableApolloUApp", false),
		DisableAppApollo:        env.Bool("ApolloDisableApolloApp", false),
	}
}

// uappZappConfig 派生「uapp 配置」使用的 zapp Apollo 参数。
func (c *ApolloConfig) uappZappConfig() *zconfig.ApolloConfig {
	backupFile := c.BackupFile
	if backupFile != "" {
		backupFile += uappBackupFileSuffix
	}

	return &zconfig.ApolloConfig{
		Address:                 c.Address,
		AppId:                   orDefault(c.UAppID, defaultUAppID),
		AccessKey:               c.AccessKey,
		AuthBasicUser:           c.AuthBasicUser,
		AuthBasicPassword:       c.AuthBasicPassword,
		Cluster:                 orDefault(c.Cluster, defaultCluster),
		AlwaysLoadFromRemote:    c.AlwaysLoadFromRemote,
		BackupFile:              backupFile,
		ApplicationDataType:     orDefault(c.ApplicationDataType, defaultApplicationDataType),
		ApplicationParseKeys:    nil,
		Namespaces:              nil,
		IgnoreNamespaceNotFound: false,
	}
}

// appZappConfig 派生「应用配置」使用的 zapp Apollo 参数。
func (c *ApolloConfig) appZappConfig(appName string) *zconfig.ApolloConfig {
	return &zconfig.ApolloConfig{
		Address:                 c.Address,
		AppId:                   orDefault(c.AppID, appName),
		AccessKey:               c.AccessKey,
		AuthBasicUser:           c.AuthBasicUser,
		AuthBasicPassword:       c.AuthBasicPassword,
		Cluster:                 orDefault(c.Cluster, defaultCluster),
		AlwaysLoadFromRemote:    c.AlwaysLoadFromRemote,
		BackupFile:              c.BackupFile,
		ApplicationDataType:     orDefault(c.ApplicationDataType, defaultApplicationDataType),
		ApplicationParseKeys:    c.ApplicationParseKeys,
		Namespaces:              c.Namespaces,
		IgnoreNamespaceNotFound: c.IgnoreNamespaceNotFound,
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
