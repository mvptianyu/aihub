/*
@Project: aihub
@Module: aihub
@File : provider_hub.go
*/
package aihub

import (
	"sync"
)

type providerHub struct {
	providers map[string]IProvider // LLM Provider Name => IProvider

	lock sync.RWMutex
}

func (h *providerHub) GetProviderList(names ...string) []IProvider {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]IProvider, 0)
	for _, name := range names {
		if tmp, ok := h.providers[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *providerHub) GetProvider(name string) IProvider {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if tmp, ok := h.providers[name]; ok {
		return tmp
	}
	return nil
}

func (h *providerHub) SetProvider(cfg *ProviderConfig) (IProvider, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if tmp, ok := h.providers[cfg.Name]; ok {
		return tmp, nil
	}

	ins, err := newProvider(cfg)
	if err != nil {
		return nil, err
	}
	h.providers[cfg.Name] = ins
	return ins, err
}
