/*
@Project: aihub
@Module: aihub
@File : provider_hub.go
*/
package aihub

import (
	"reflect"
	"strings"
	"sync"
)

type middlewareHub struct {
	middlewares map[string]IMiddleware // Middleware Name => IMiddleware

	lock sync.RWMutex
}

func (h *middlewareHub) GetMiddleware(names ...string) []IMiddleware {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]IMiddleware, 0)
	for _, name := range names {
		if tmp, ok := h.middlewares[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *middlewareHub) SetMiddleware(objs ...IMiddleware) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, obj := range objs {
		rv := reflect.TypeOf(obj)
		// 获取结构名
		structName := rv.Elem().Name()
		splits := strings.Split(structName, ".")
		fixName := splits[len(splits)-1] // 去除结构名中的包路径
		h.middlewares[fixName] = obj
	}
	return nil
}

func (h *middlewareHub) DelMiddleware(names ...string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, name := range names {
		delete(h.middlewares, name)
	}
	return nil
}
