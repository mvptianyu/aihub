package aihub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type llmHub struct {
	llms map[string]ILLM // Name => ILLM

	lock sync.RWMutex
}

func (h *llmHub) GetAllNameList() []string {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]string, 0)
	for name, _ := range h.llms {
		ret = append(ret, name)
	}
	return ret
}

func (h *llmHub) GetLLMList(names ...string) []ILLM {
	h.lock.RLock()
	defer h.lock.RUnlock()

	ret := make([]ILLM, 0)
	for _, name := range names {
		if tmp, ok := h.llms[name]; ok {
			ret = append(ret, tmp)
		}
	}
	return ret
}

func (h *llmHub) GetLLM(name string) ILLM {
	h.lock.RLock()
	defer h.lock.RUnlock()
	if tmp, ok := h.llms[name]; ok {
		return tmp
	}
	return nil
}

func (h *llmHub) DelLLM(name string) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.llms, name)
	return nil
}

func (h *llmHub) SetLLM(cfg *LLMConfig) (ILLM, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if _, ok := h.llms[cfg.Name]; ok {
		delete(h.llms, cfg.Name) // 删除旧的
	}

	ins, err := newLLM(cfg)
	if err != nil {
		return nil, err
	}
	h.llms[cfg.Name] = ins
	return ins, err
}

func (h *llmHub) SetLLMByYamlData(yamlData []byte) (ILLM, error) {
	cfg, err := YamlDataToLLMConfig(yamlData)
	if err != nil {
		return nil, err
	}
	return h.SetLLM(cfg)
}

func (h *llmHub) SetLLMByYamlFile(yamlFile string) (ILLM, error) {
	// 读取 YAML 文件内容
	yamlData, err := os.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		fmt.Printf("Error reading YAML file: %s => %v\n", yamlFile, err)
		return nil, err
	}
	return h.SetLLMByYamlData(yamlData)
}
