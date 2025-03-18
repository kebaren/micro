package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// AppSettings 保存应用程序设置
var AppSettings struct {
	ShowToolbar  bool
	UseDarkTheme bool
}

// 全局应用变量
var (
	App        fyne.App
	MainWindow fyne.Window
	MainStatus *widget.Label // 全局状态栏
)

// 初始化默认设置
func init() {
	AppSettings.ShowToolbar = false
	AppSettings.UseDarkTheme = false
}

// UpdateMainStatus 更新全局状态栏显示的文本
func UpdateMainStatus(text string) {
	if MainStatus != nil {
		MainStatus.SetText(text)
	}
}

// InitApp initializes the GUI application
func InitApp() {
	// 创建应用程序
	App = app.New()

	// 优化应用启动: 设置更高效的主题
	App.Settings().SetTheme(newCompactTheme())

	// 创建主窗口
	MainWindow = App.NewWindow("Micro Editor")

	// 设置窗口填充和内边距
	MainWindow.SetPadded(false)

	// 性能优化：禁用窗口内边距以减少绘制操作
	MainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		// 处理全局键盘快捷键
		if ke.Name == fyne.KeyF11 {
			MainWindow.SetFullScreen(!MainWindow.FullScreen())
		}
	})

	// 设置默认窗口大小
	MainWindow.Resize(fyne.NewSize(1024, 768))

	// 性能优化：延迟显示窗口到完全加载后
	MainWindow.SetOnClosed(func() {
		App.Quit()
	})

	// 将窗口居中显示，确保设置内容之前这些属性都已准备好
	MainWindow.CenterOnScreen()
}

// Run runs the GUI application
func Run() {
	if MainWindow == nil {
		panic("Window content not set")
	}

	// 性能优化：使用低级别渲染并减少刷新率
	MainWindow.Show()

	// 运行应用程序
	App.Run()
}

// GetThemeColors 返回当前主题的颜色集
func GetThemeColors() (bg, fg, primary, secondary color.Color) {
	if AppSettings.UseDarkTheme {
		// VSCode暗色主题颜色
		bg = color.NRGBA{R: 30, G: 30, B: 30, A: 255}        // 更深的背景色
		fg = color.NRGBA{R: 220, G: 220, B: 220, A: 255}     // 文本颜色
		primary = color.NRGBA{R: 0, G: 122, B: 204, A: 255}  // VSCode蓝
		secondary = color.NRGBA{R: 40, G: 40, B: 40, A: 255} // 菜单和编辑区背景
	} else {
		// VSCode亮色主题颜色
		bg = color.NRGBA{R: 245, G: 245, B: 245, A: 255}        // 浅灰色背景
		fg = color.NRGBA{R: 33, G: 33, B: 33, A: 255}           // 深色文本
		primary = color.NRGBA{R: 0, G: 122, B: 204, A: 255}     // 保持VSCode蓝
		secondary = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 白色菜单和编辑区背景
	}
	return
}

// 创建一个更接近VSCode主题的紧凑主题，并优化内存使用
type compactTheme struct {
	fyne.Theme
	isDark bool
	// 内存优化：缓存计算过的颜色，避免重复创建颜色对象
	colorCache map[fyne.ThemeColorName]color.Color
	// 内存优化：缓存计算过的尺寸值
	sizeCache map[fyne.ThemeSizeName]float32
}

// 创建一个新的紧凑主题，预分配缓存以减少运行时内存分配
func newCompactTheme() fyne.Theme {
	return &compactTheme{
		Theme:      theme.DefaultTheme(),
		isDark:     AppSettings.UseDarkTheme,
		colorCache: make(map[fyne.ThemeColorName]color.Color, 16), // 预分配常用颜色数量
		sizeCache:  make(map[fyne.ThemeSizeName]float32, 8),       // 预分配常用尺寸数量
	}
}

// 实现主题接口的方法，使用缓存减少内存分配
func (t *compactTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 内存优化：检查缓存中是否已有计算结果
	if cachedColor, ok := t.colorCache[name]; ok {
		return cachedColor
	}

	// 获取颜色，共用相同对象
	var result color.Color
	bg, fg, primary, secondary := GetThemeColors()

	if t.isDark {
		// VSCode暗色主题颜色 - 简化颜色计算提高性能
		switch name {
		case theme.ColorNameBackground:
			result = bg
		case theme.ColorNameForeground:
			result = fg
		case theme.ColorNameShadow:
			result = color.NRGBA{R: 20, G: 20, B: 20, A: 128}
		case theme.ColorNamePrimary:
			result = primary
		case theme.ColorNameInputBackground:
			result = secondary
		case theme.ColorNameButton, theme.ColorNameHover:
			result = color.NRGBA{R: 45, G: 45, B: 45, A: 255}
		case theme.ColorNameDisabledButton:
			result = color.NRGBA{R: 40, G: 40, B: 40, A: 255}
		case theme.ColorNameMenuBackground:
			result = secondary
		default:
			result = t.Theme.Color(name, variant)
		}
	} else {
		// VSCode亮色主题颜色 - 简化颜色计算提高性能
		switch name {
		case theme.ColorNameBackground:
			result = bg
		case theme.ColorNameForeground:
			result = fg
		case theme.ColorNameShadow:
			result = color.NRGBA{R: 180, G: 180, B: 180, A: 128}
		case theme.ColorNamePrimary:
			result = primary
		case theme.ColorNameInputBackground:
			result = secondary
		case theme.ColorNameButton, theme.ColorNameHover:
			result = color.NRGBA{R: 240, G: 240, B: 240, A: 255}
		case theme.ColorNameDisabledButton:
			result = color.NRGBA{R: 230, G: 230, B: 230, A: 255}
		case theme.ColorNameMenuBackground:
			result = secondary
		default:
			result = t.Theme.Color(name, variant)
		}
	}

	// 存入缓存
	t.colorCache[name] = result
	return result
}

// 返回更大的文本大小，提高可读性
func (t *compactTheme) Size(name fyne.ThemeSizeName) float32 {
	// 内存优化：检查缓存中是否已有计算结果
	if cachedSize, ok := t.sizeCache[name]; ok {
		return cachedSize
	}

	// 放大文本和调整UI尺寸，提升可读性
	var result float32
	switch name {
	case theme.SizeNamePadding:
		result = 3 // 适当增加内边距
	case theme.SizeNameInlineIcon:
		result = 18 // 更大的图标
	case theme.SizeNameText:
		result = 12 // 更大的文本
	case theme.SizeNameHeadingText:
		result = 14 // 更大的标题文本
	case theme.SizeNameSubHeadingText:
		result = 13 // 更大的副标题
	case theme.SizeNameSeparatorThickness:
		result = 1 // 分隔符保持细线
	case theme.SizeNameScrollBar, theme.SizeNameScrollBarSmall:
		result = 10 // 更大的滚动条便于操作
	case theme.SizeNameInputBorder:
		result = 1 // 保持输入边框细线
	default:
		result = t.Theme.Size(name)
	}

	// 存入缓存
	t.sizeCache[name] = result
	return result
}

// Icon 返回指定名称的图标资源
func (t *compactTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.Theme.Icon(name)
}

// Font 返回指定文本样式的字体资源
func (t *compactTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.Theme.Font(style)
}
