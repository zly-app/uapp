package uapp

import (
	"github.com/zly-app/cache/v2"
	"github.com/zly-app/component/es7"
	"github.com/zly-app/component/mongo"
	pulsar_producer "github.com/zly-app/component/pulsar-producer"
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/sqlx"
	"github.com/zly-app/component/xorm"
	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/core"
)

type IComponent interface {
	core.IComponent

	es7.IES7
	mongo.IMongoCreator
	redis.IRedisCreator
	sqlx.ISqlx
	xorm.IXormCreator
	cache.ICacheCreator
	pulsar_producer.IPulsarProducerCreator
}

type Component struct {
	core.IComponent

	es7.IES7
	mongo.IMongoCreator
	redis.IRedisCreator
	sqlx.ISqlx
	xorm.IXormCreator
	pulsar_producer.IPulsarProducerCreator
	cache.ICacheCreator
}

func (c *Component) Close() {
	c.IComponent.Close()

	c.IES7.Close()
	c.IMongoCreator.Close()
	c.IRedisCreator.Close()
	c.ISqlx.Close()
	c.IXormCreator.Close()
	c.IPulsarProducerCreator.Close()
	c.ICacheCreator.Close()
}

func makeCustomComponent(app core.IApp, c core.IComponent) core.IComponent {
	return &Component{
		IComponent: c,

		IES7:                   es7.NewES7(app),
		IMongoCreator:          mongo.NewMongoCreator(app),
		IRedisCreator:          redis.NewRedisCreator(app),
		ISqlx:                  sqlx.NewSqlx(app),
		IXormCreator:           xorm.NewXormCreator(app),
		IPulsarProducerCreator: pulsar_producer.NewProducerCreator(app),
		ICacheCreator:          cache.NewCacheCreator(app),
	}
}

// 获取组件
func GetComponent() IComponent {
	c := component.GetComponent()
	return c.(IComponent)
}
