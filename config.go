package uapp

import (
	"flag"
	"os"

	"github.com/spf13/cast"
	"github.com/zly-app/zapp/pkg/utils"
)

const (
	ApolloAddress                 = "ApolloAddress"                 // apollo-api地址, 多个地址用英文逗号连接, 如果不存在则不会使用apollo
	ApolloUAppID                  = "ApolloUAppID"                  // uapp 应用名
	ApolloAppId                   = "ApolloAppId"                   // 当前应用名, 应用要覆盖 uapp 的配置
	ApolloDisableApolloUApp       = "ApolloDisableApolloUApp"       // uapp不从apollo中获取uapp配置, 不会影响`应用配置`的获取
	ApolloDisableApolloApp        = "ApolloDisableApolloApp"        // uapp不从apollo中获取`应用配置`
	ApolloAccessKey               = "ApolloAccessKey"               // 验证key, 优先级高于基础认证
	ApolloAuthBasicUser           = "ApolloAuthBasicUser"           // 基础认证用户名, 可用于nginx的基础认证扩展
	ApolloAuthBasicPassword       = "ApolloAuthBasicPassword"       // 基础认证密码
	ApolloCluster                 = "ApolloCluster"                 // 集群名
	ApolloAlwaysLoadFromRemote    = "ApolloAlwaysLoadFromRemote"    // 总是从远程获取, 在远程加载失败时不会从备份文件加载
	ApolloBackupFile              = "ApolloBackupFile"              // 备份文件名
	ApolloApplicationDataType     = "ApolloApplicationDataType"     // application命名空间下key的值的数据类型, 支持yaml,yml,toml,json
	ApolloApplicationParseKeys    = "ApolloApplicationParseKeys"    // application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/services)会被解析
	ApolloNamespaces              = "ApolloNamespaces"              // 其他自定义命名空间
	ApolloIgnoreNamespaceNotFound = "ApolloIgnoreNamespaceNotFound" // 忽略其他自定义命名空间不存在
)

type confList struct {
	ApolloAddress                 *string
	ApolloUAppID                  *string
	ApolloAppId                   *string
	ApolloDisableApolloUApp       *bool
	ApolloDisableApolloApp        *bool
	ApolloAccessKey               *string
	ApolloAuthBasicUser           *string
	ApolloAuthBasicPassword       *string
	ApolloCluster                 *string
	ApolloAlwaysLoadFromRemote    *bool
	ApolloBackupFile              *string
	ApolloApplicationDataType     *string
	ApolloApplicationParseKeys    *string
	ApolloNamespaces              *string
	ApolloIgnoreNamespaceNotFound *bool
}

var cl = &confList{}

func initConfig() {
	cl.ApolloAddress = flag.String(ApolloAddress, os.Getenv(ApolloAddress), "apollo-api地址, 多个地址用英文逗号连接, 如果不存在则不会使用apollo")
	cl.ApolloUAppID = flag.String(ApolloUAppID, utils.Ternary.Or(os.Getenv(ApolloUAppID), "uapp").(string), "uapp 应用名")
	cl.ApolloAppId = flag.String(ApolloAppId, os.Getenv(ApolloAppId), "当前应用名, 应用要覆盖 uapp 的配置")
	cl.ApolloDisableApolloUApp = flag.Bool(ApolloDisableApolloUApp, cast.ToBool(os.Getenv(ApolloDisableApolloUApp)), "uapp不从apollo中获取uapp配置, 不会影响`应用配置`的获取")
	cl.ApolloDisableApolloApp = flag.Bool(ApolloDisableApolloApp, cast.ToBool(os.Getenv(ApolloAlwaysLoadFromRemote)), "uapp不从apollo中获取`应用配置`")
	cl.ApolloAccessKey = flag.String(ApolloAccessKey, os.Getenv(ApolloAccessKey), "验证key, 优先级高于基础认证")
	cl.ApolloAuthBasicUser = flag.String(ApolloAuthBasicUser, os.Getenv(ApolloAuthBasicUser), "基础认证用户名, 可用于nginx的基础认证扩展")
	cl.ApolloAuthBasicPassword = flag.String(ApolloAuthBasicPassword, os.Getenv(ApolloAuthBasicPassword), "基础认证密码")
	cl.ApolloCluster = flag.String(ApolloCluster, utils.Ternary.Or(os.Getenv(ApolloCluster), "default").(string), "集群名")
	cl.ApolloAlwaysLoadFromRemote = flag.Bool(ApolloAlwaysLoadFromRemote, cast.ToBool(os.Getenv(ApolloAlwaysLoadFromRemote)), "总是从远程获取, 在远程加载失败时不会从备份文件加载")
	cl.ApolloBackupFile = flag.String(ApolloBackupFile, os.Getenv(ApolloBackupFile), "备份文件名")
	cl.ApolloApplicationDataType = flag.String(ApolloApplicationDataType, utils.Ternary.Or(os.Getenv(ApolloApplicationDataType), "yaml").(string), "application命名空间下key的值的数据类型, 支持yaml,yml,toml,json")
	cl.ApolloApplicationParseKeys = flag.String(ApolloApplicationParseKeys, os.Getenv(ApolloApplicationParseKeys), "application命名空间下哪些key数据会被解析, 无论如何默认的key(frame/components/plugins/services)会被解析")
	cl.ApolloNamespaces = flag.String(ApolloNamespaces, os.Getenv(ApolloNamespaces), "其他自定义命名空间")
	cl.ApolloIgnoreNamespaceNotFound = flag.Bool(ApolloIgnoreNamespaceNotFound, cast.ToBool(os.Getenv(ApolloIgnoreNamespaceNotFound)), "忽略其他自定义命名空间不存在")
}

func getConfig() *confList {
	flag.Parse()
	return cl
}
