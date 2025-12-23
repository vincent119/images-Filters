package filter

import (
	"fmt"
	"image"

	"github.com/vincent119/images-filters/pkg/logger"
)

// FilterSpec 濾鏡規格（從 URL 解析）
type FilterSpec struct {
	Name   string
	Params []string
}

// Pipeline 濾鏡管線
// 依序執行多個濾鏡
type Pipeline struct {
	registry *Registry
	specs    []FilterSpec
}

// NewPipeline 建立濾鏡管線
func NewPipeline(registry *Registry) *Pipeline {
	if registry == nil {
		registry = DefaultRegistry()
	}
	return &Pipeline{
		registry: registry,
		specs:    make([]FilterSpec, 0),
	}
}

// Add 加入濾鏡規格
func (p *Pipeline) Add(spec FilterSpec) *Pipeline {
	p.specs = append(p.specs, spec)
	return p
}

// AddMany 加入多個濾鏡規格
func (p *Pipeline) AddMany(specs []FilterSpec) *Pipeline {
	p.specs = append(p.specs, specs...)
	return p
}

// Clear 清空管線
func (p *Pipeline) Clear() *Pipeline {
	p.specs = p.specs[:0]
	return p
}

// Count 取得管線中的濾鏡數量
func (p *Pipeline) Count() int {
	return len(p.specs)
}

// Apply 執行管線
// 依序對圖片應用所有濾鏡
func (p *Pipeline) Apply(img image.Image) (image.Image, error) {
	if len(p.specs) == 0 {
		return img, nil
	}

	current := img

	for i, spec := range p.specs {
		// 取得濾鏡
		filter, exists := p.registry.Get(spec.Name)
		if !exists {
			logger.Debug("filter not found, skipping",
				logger.String("filter", spec.Name),
				logger.Int("index", i),
			)
			continue
		}

		// 應用濾鏡
		result, err := filter.Apply(current, spec.Params)
		if err != nil {
			logger.Debug("filter apply failed",
				logger.String("filter", spec.Name),
				logger.Err(err),
			)
			return nil, fmt.Errorf("filter '%s' failed: %w", spec.Name, err)
		}

		logger.Debug("filter applied",
			logger.String("filter", spec.Name),
			logger.Int("index", i),
		)

		current = result
	}

	return current, nil
}

// ApplyWithCallback 執行管線（帶回調）
// 每個濾鏡執行後會調用回調函數
func (p *Pipeline) ApplyWithCallback(img image.Image, callback func(name string, result image.Image)) (image.Image, error) {
	if len(p.specs) == 0 {
		return img, nil
	}

	current := img

	for _, spec := range p.specs {
		filter, exists := p.registry.Get(spec.Name)
		if !exists {
			continue
		}

		result, err := filter.Apply(current, spec.Params)
		if err != nil {
			return nil, fmt.Errorf("filter '%s' failed: %w", spec.Name, err)
		}

		if callback != nil {
			callback(spec.Name, result)
		}

		current = result
	}

	return current, nil
}
