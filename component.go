package uapp

import (
	"github.com/zly-app/cache"
	"github.com/zly-app/component/es7"
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/sqlx"
	"github.com/zly-app/component/xorm"
	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/core"
)

type IComponent interface {
	core.IComponent

	es7.IES7
	redis.IRedisCreator
	sqlx.ISqlx
	xorm.IXormCreator
	cache.ICacheCreator
}

type Component struct {
	core.IComponent

	es7.IES7
	redis.IRedisCreator
	sqlx.ISqlx
	xorm.IXormCreator
	cache.ICacheCreator
}

func (c *Component) Close() {
	c.IComponent.Close()

	c.IES7.Close()
	c.IRedisCreator.Close()
	c.ISqlx.Close()
	c.IXormCreator.Close()
	c.ICacheCreator.Close()
}

func makeCustomComponent(app core.IApp) core.IComponent {
	return &Component{
		IComponent: app.GetComponent(),

		IES7:          es7.NewES7(app),
		IRedisCreator: redis.NewRedisCreator(app),
		ISqlx:         sqlx.NewSqlx(app),
		IXormCreator:  xorm.NewXormCreator(app),
		ICacheCreator: cache.NewCacheCreator(app),
	}
}

func GetComponent() IComponent {
	c := component.GetComponent()
	return c.(IComponent)
}
