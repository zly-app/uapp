package uapp

import (
	"github.com/zly-app/component/redis"
	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/core"
)

type IComponent interface {
	core.IComponent
	redis.IRedisCreator
}

type Component struct {
	core.IComponent
	redis.IRedisCreator
}

func (c *Component) Close() {
	c.IComponent.Close()
	c.IRedisCreator.Close()
}

func makeCustomComponent(app core.IApp) core.IComponent {
	return &Component{
		IComponent:    app.GetComponent(),
		IRedisCreator: redis.NewRedisCreator(app),
	}
}

func GetComponent() IComponent {
	c := component.GetComponent()
	return c.(IComponent)
}
