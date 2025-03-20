package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

// 全局应用设置
var AppSettings = AppSettingsData{
	Theme:      "dark",      // 默认使用深色主题
	FontSize:   16,          // 增加字体大小
	LineHeight: 1.6,         // 增加行高
	FontFamily: "monospace", // 默认等宽字体
}

// AppSettingsData 存储应用程序设置
type AppSettingsData struct {
	Theme      string
	FontSize   float32
	LineHeight float32
	FontFamily string
}

// App 表示主应用程序
type App struct {
	fyneApp    fyne.App
	settings   AppSettingsData
	mainWindow fyne.Window
}

// NewApp 创建新的应用程序实例
func NewApp() *App {
	app := &App{
		fyneApp: app.New(),
		settings: AppSettingsData{
			Theme:      "dark",
			FontSize:   16,
			LineHeight: 1.6,
			FontFamily: "monospace",
		},
	}
	// 应用主题
	app.applySettings()
	return app
}

// SetTheme 设置应用程序主题
func (a *App) SetTheme(name string) {
	a.settings.Theme = name
	// 更新全局设置
	AppSettings.Theme = name
	a.applySettings()
}

// SetFontSize 设置字体大小
func (a *App) SetFontSize(size float32) {
	a.settings.FontSize = size
	// 更新全局设置
	AppSettings.FontSize = size
	a.applySettings()
}

// SetFontFamily 设置字体系列
func (a *App) SetFontFamily(family string) {
	a.settings.FontFamily = family
	// 更新全局设置
	AppSettings.FontFamily = family
	a.applySettings()
}

// applySettings 应用所有设置
func (a *App) applySettings() {
	// 应用主题
	if a.settings.Theme == "dark" {
		fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
	} else {
		fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	}

	// 其他设置可以在此应用
	// 注意：Fyne目前不直接支持修改字体大小和系列，需要自定义主题
}

// getTheme 获取当前主题
func (a *App) getTheme() fyne.Theme {
	if a.settings.Theme == "dark" {
		return theme.DarkTheme()
	}
	return theme.LightTheme()
}

// NewWindow 创建新窗口
func (a *App) NewWindow(title string) fyne.Window {
	a.mainWindow = a.fyneApp.NewWindow(title)
	// 设置窗口默认大小更合理
	a.mainWindow.Resize(fyne.NewSize(1024, 768))
	return a.mainWindow
}

// Quit 退出应用程序
func (a *App) Quit() {
	a.fyneApp.Quit()
}
