package aihub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type providerHub struct {
	providers map[string]IProvider // LLM Provider Name => IProvider

	lock sync.RWMutex
}

func (h *providerHub) GetAllNameList() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range h.providers {
		ret = append(ret, name)
	}
	return ret
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

func (h *providerHub) DelProvider(name string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.providers, name)
	return nil
}

func (h *providerHub) SetProvider(cfg *ProviderConfig) (IProvider, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.providers[cfg.Name]; ok {
		delete(h.providers, cfg.Name) // 删除旧的
	}

	ins, err := newProvider(cfg)
	if err != nil {
		return nil, err
	}
	h.providers[cfg.Name] = ins
	return ins, err
}

func (h *providerHub) SetProviderByYamlData(yamlData []byte) (IProvider, error) {
	cfg, err := YamlDataToProviderConfig(yamlData)
	if err != nil {
		return nil, err
	}
	return h.SetProvider(cfg)
}

func (h *providerHub) SetProviderByYamlFile(yamlFile string) (IProvider, error) {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		fmt.Printf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil, err
	}
	return h.SetProviderByYamlData(yamlData)
}
