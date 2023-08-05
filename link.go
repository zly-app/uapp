package uapp

import (
	"github.com/jmoiron/sqlx"
	elastic7 "github.com/olivere/elastic/v7"
	"github.com/zly-app/cache"
	"github.com/zly-app/component/mongo"
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/xorm"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// 获取默认es7客户端
func GetDefES7() *elastic7.Client {
	return GetComponent().GetDefES7()
}

// 获取默认mongo客户端
func GetDefMongo() *mongo.Client {
	return GetComponent().GetDefMongo()
}

// 获取默认redis客户端
func GetDefRedis() redis.UniversalClient {
	return GetComponent().GetDefRedis()
}

// 获取默认sqlx客户端
func GetDefSqlx() *sqlx.DB {
	return GetComponent().GetDefSqlx()
}

// 获取默认xorm客户端
func GetDefXorm() *xorm.Engine {
	return GetComponent().GetDefXorm()
}

// 获取默认缓存
func GetDefCache() cache.ICache {
	return GetComponent().GetDefCache()
}

// 获取默认协程池
func GetDefGPool() core.IGPool {
	return GetComponent().GetGPool(consts.DefaultComponentName)
}

// 获取消息总线
func GetMsgbus() core.IMsgbus {
	return GetComponent()
}

// 获取日志工具
func GetLogger() core.ILogger {
	return GetComponent()
}

// 获取app
func GetApp() core.IApp {
	return GetComponent().App()
}

// 获取配置
func GetConfig() core.IConfig {
	return config.Conf
}
