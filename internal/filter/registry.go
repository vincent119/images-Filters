package filter

import (
	"fmt"
	"sync"
)

// Registry 濾鏡註冊表
// 管理所有已註冊的濾鏡
type Registry struct {
	mu      sync.RWMutex
	filters map[string]Filter
}

// NewRegistry 建立新的濾鏡註冊表
func NewRegistry() *Registry {
	return &Registry{
		filters: make(map[string]Filter),
	}
}

// Register 註冊濾鏡
func (r *Registry) Register(filter Filter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := filter.Name()
	if _, exists := r.filters[name]; exists {
		return fmt.Errorf("filter already registered: %s", name)
	}

	r.filters[name] = filter
	return nil
}

// MustRegister 註冊濾鏡，失敗時 panic
func (r *Registry) MustRegister(filter Filter) {
	if err := r.Register(filter); err != nil {
		panic(err)
	}
}

// Get 取得濾鏡
func (r *Registry) Get(name string) (Filter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filter, exists := r.filters[name]
	return filter, exists
}

// List 列出所有已註冊的濾鏡名稱
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.filters))
	for name := range r.filters {
		names = append(names, name)
	}
	return names
}

// Count 取得已註冊濾鏡數量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.filters)
}

// 全域預設 Registry
var defaultRegistry = NewRegistry()

// DefaultRegistry 取得全域預設 Registry
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// Register 註冊濾鏡到全域預設 Registry
func Register(filter Filter) error {
	return defaultRegistry.Register(filter)
}

// MustRegister 註冊濾鏡到全域預設 Registry，失敗時 panic
func MustRegister(filter Filter) {
	defaultRegistry.MustRegister(filter)
}

// Get 從全域預設 Registry 取得濾鏡
func Get(name string) (Filter, bool) {
	return defaultRegistry.Get(name)
}
