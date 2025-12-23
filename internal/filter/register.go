package filter

// RegisterDefaultFilters 註冊所有預設濾鏡
func RegisterDefaultFilters(r *Registry) {
	// 基本濾鏡
	r.MustRegister(NewBlurFilter())
	r.MustRegister(NewGrayscaleFilter())
	r.MustRegister(NewBrightnessFilter())
	r.MustRegister(NewContrastFilter())
	r.MustRegister(NewSaturationFilter())
	r.MustRegister(NewSharpenFilter())
	r.MustRegister(NewInvertFilter())
	r.MustRegister(NewNoOpFilter())

	// 顏色處理濾鏡
	r.MustRegister(NewRGBFilter())
	r.MustRegister(NewSepiaFilter())
	r.MustRegister(NewEqualizeFilter())
	r.MustRegister(NewGammaFilter())
	r.MustRegister(NewHueFilter())

	// 特效濾鏡
	r.MustRegister(NewRotateFilter())
	r.MustRegister(NewRoundCornersFilter())
	r.MustRegister(NewNoiseFilter())
	r.MustRegister(NewFlipHFilter())
	r.MustRegister(NewFlipVFilter())
	r.MustRegister(NewPixelateFilter())

	// 輸出控制濾鏡
	r.MustRegister(NewQualityFilter())
	r.MustRegister(NewFormatFilter())
	r.MustRegister(NewStripExifFilter())
	r.MustRegister(NewStripICCFilter())
	r.MustRegister(NewAutoOrientFilter())

	// 浮水印濾鏡
	r.MustRegister(NewWatermarkFilter())
}

// init 自動註冊到全域 Registry
func init() {
	RegisterDefaultFilters(DefaultRegistry())
}
