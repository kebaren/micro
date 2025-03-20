package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	ActivityBarWidth = 40
	SideBarWidth     = 220
	StatusBarHeight  = 20
	MinSideBarWidth  = 180
	MaxSideBarWidth  = 400
	TabBarHeight     = 28
)

// MainGUI 表示主界面
type MainGUI struct {
	app              *App
	activityBar      *fyne.Container
	sideBar          *fyne.Container
	resizableSidebar *ResizeableSidebar
	editor           *EditorView
	editorArea       *fyne.Container
	bottomPanel      *BottomPanel
	statusBar        *fyne.Container
	content          *fyne.Container
	fileBrowser      *FileBrowser
	sideBarWidth     float32
	sideBarVisible   bool
	activeSideTab    int
	windows          []*fyne.Window
}

// NewMainGUI 创建新的主界面
func NewMainGUI(app *App) *MainGUI {
	gui := &MainGUI{
		app:            app,
		sideBarWidth:   SideBarWidth,
		sideBarVisible: true,
		activeSideTab:  0,
	}

	// 创建活动栏
	gui.activityBar = gui.createActivityBar()

	// 创建文件浏览器
	gui.fileBrowser = NewFileBrowser("", gui.openFile)

	// 创建侧边栏
	gui.sideBar = gui.createSideBar()

	// 创建可调整大小的侧边栏
	gui.resizableSidebar = NewResizeableSidebar(gui.sideBar, gui.sideBarWidth, MinSideBarWidth, MaxSideBarWidth, func(width float32) {
		// 当宽度改变时的回调
		gui.sideBarWidth = width

		// 调整侧边栏大小
		gui.sideBar.Resize(fyne.NewSize(width, gui.app.mainWindow.Canvas().Size().Height))

		// 更新整个布局
		gui.updateLayout()

		// 重要：强制更新整个窗口内容
		if gui.app != nil && gui.app.mainWindow != nil {
			gui.app.mainWindow.SetContent(gui.content)
			gui.app.mainWindow.Canvas().Refresh(gui.content)
		}
	})

	// 创建编辑器
	gui.editor = NewEditorView()

	// 创建编辑器区域
	gui.editorArea = gui.createEditorArea()

	// 创建底部面板
	gui.bottomPanel = NewBottomPanel()

	// 创建状态栏
	gui.statusBar = gui.createStatusBar()

	// 创建主布局
	gui.updateLayout()

	// 确保布局在窗口大小变化时更新
	app.mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		// 处理按键事件
	})

	return gui
}

// updateLayout 更新布局
func (g *MainGUI) updateLayout() {
	// 获取窗口尺寸
	windowWidth := g.app.mainWindow.Canvas().Size().Width
	windowHeight := g.app.mainWindow.Canvas().Size().Height

	var leftSide fyne.CanvasObject

	if g.sideBarVisible {
		// 应用新的宽度到侧边栏和可调整大小的侧边栏
		g.sideBar.Resize(fyne.NewSize(g.sideBarWidth, windowHeight))
		g.resizableSidebar.Resize(fyne.NewSize(g.sideBarWidth, windowHeight))

		leftSide = container.NewHBox(
			g.activityBar,
			g.resizableSidebar,
		)
	} else {
		leftSide = g.activityBar
	}

	// 计算编辑器区域宽度 - 确保填充所有可用空间
	editorWidth := windowWidth - ActivityBarWidth
	if g.sideBarVisible {
		editorWidth -= g.sideBarWidth
	}

	// 使用Max容器确保编辑器区域填充所有可用空间
	g.editorArea.Resize(fyne.NewSize(editorWidth, windowHeight-g.bottomPanel.GetContent().Size().Height-StatusBarHeight))
	editorContainer := container.NewMax(g.editorArea)

	// 创建底部区域，根据底部面板的可见性决定是否显示
	var bottomArea fyne.CanvasObject
	if g.bottomPanel.IsVisible() {
		// 调整底部面板宽度
		g.bottomPanel.GetContent().Resize(fyne.NewSize(editorWidth, g.bottomPanel.GetContent().Size().Height))
		bottomArea = container.NewMax(g.bottomPanel.GetContent())
	} else {
		// 隐藏底部面板，使用空容器
		bottomArea = container.NewMax()
	}

	// 创建主布局
	mainContent := container.NewBorder(
		nil,        // top
		bottomArea, // bottom
		nil, nil,   // left, right
		editorContainer, // center
	)

	// 创建完整布局
	g.content = container.NewBorder(
		nil,         // top
		g.statusBar, // bottom
		leftSide,    // left
		nil,         // right
		mainContent, // center
	)

	// 重要：立即设置主窗口内容以应用更新
	if g.app != nil && g.app.mainWindow != nil {
		g.app.mainWindow.SetContent(g.content)
		// 强制刷新以确保布局更新
		g.app.mainWindow.Canvas().Refresh(g.content)
	}
}

// createResizeHandle 创建侧边栏宽度调整控件
func (g *MainGUI) createResizeHandle() fyne.CanvasObject {
	handle := widget.NewLabel("")
	handle.Resize(fyne.NewSize(4, 1000)) // 垂直调整条

	// 自定义渲染器
	handle.ExtendBaseWidget(handle)

	// 在这里我们需要实现鼠标拖拽逻辑
	// 由于Fyne的标准组件不直接支持拖拽调整大小，这里提供一个示意
	// 实际实现时可能需要自定义组件或使用其他方法

	return container.NewHBox(
		widget.NewSeparator(),
	)
}

// GetContent 获取主界面内容
func (g *MainGUI) GetContent() fyne.CanvasObject {
	return g.content
}

// toggleSideBar 切换侧边栏显示状态
func (g *MainGUI) toggleSideBar(tabIndex int) {
	if g.activeSideTab == tabIndex && g.sideBarVisible {
		// 如果点击当前活动的标签页，则隐藏侧边栏
		g.sideBarVisible = false
	} else {
		// 否则显示侧边栏并切换到该标签页
		g.sideBarVisible = true
		g.activeSideTab = tabIndex
		g.updateSideBarContent()
	}

	// 更新布局
	g.updateLayout()

	// 刷新内容
	if g.app != nil && g.app.mainWindow != nil {
		g.app.mainWindow.SetContent(g.content)
	}
}

// updateSideBarContent 更新侧边栏内容
func (g *MainGUI) updateSideBarContent() {
	// 基于activeSideTab更新侧边栏内容
	switch g.activeSideTab {
	case 0: // 文件浏览器
		header := container.NewBorder(
			nil, nil,
			widget.NewLabel("资源管理器"),
			container.NewHBox(
				widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
					// 创建新文件
					g.createNewFile()
				}),
				widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
					// 打开文件夹
					g.openFolder()
				}),
				widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
					// 刷新
					g.refreshFileBrowser()
				}),
			),
		)

		g.sideBar.Objects = []fyne.CanvasObject{
			container.NewBorder(
				header,
				nil, nil, nil,
				container.NewVBox(), // 清空内容区域
			),
		}
	case 1: // 搜索
		g.sideBar.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					widget.NewLabel("搜索"),
					widget.NewEntry(),
				),
				nil, nil, nil,
				container.NewVBox(), // 清空内容区域
			),
		}
	case 2: // 源代码管理
		g.sideBar.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					widget.NewLabel("源代码管理"),
				),
				nil, nil, nil,
				container.NewVBox(), // 清空内容区域
			),
		}
	case 3: // 调试
		g.sideBar.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					widget.NewLabel("运行和调试"),
				),
				nil, nil, nil,
				container.NewVBox(), // 清空内容区域
			),
		}
	case 4: // 扩展
		g.sideBar.Objects = []fyne.CanvasObject{
			container.NewBorder(
				container.NewVBox(
					widget.NewLabel("扩展"),
					widget.NewEntry(),
				),
				nil, nil, nil,
				container.NewVBox(), // 清空内容区域
			),
		}
	}

	// 应用新的大小
	g.sideBar.Resize(fyne.NewSize(g.sideBarWidth, g.app.mainWindow.Canvas().Size().Height))
}

// createActivityBar 创建活动栏
func (g *MainGUI) createActivityBar() *fyne.Container {
	// 创建活动栏按钮 - 使用更小的尺寸
	explorerBtn := widget.NewButtonWithIcon("", theme.FolderIcon(), func() {
		g.toggleSideBar(0)
	})
	explorerBtn.Importance = widget.LowImportance

	searchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		g.toggleSideBar(1)
	})
	searchBtn.Importance = widget.LowImportance

	sourceControlBtn := widget.NewButtonWithIcon("", theme.StorageIcon(), func() {
		g.toggleSideBar(2)
	})
	sourceControlBtn.Importance = widget.LowImportance

	runDebugBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		g.toggleSideBar(3)
	})
	runDebugBtn.Importance = widget.LowImportance

	extensionsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		g.toggleSideBar(4)
	})
	extensionsBtn.Importance = widget.LowImportance

	// 添加设置按钮
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		// 在这里添加设置菜单处理
		popup := widget.NewModalPopUp(
			container.NewVBox(
				widget.NewButton("设置", nil),
				widget.NewButton("键盘快捷键", nil),
				widget.NewButton("用户片段", nil),
				widget.NewButton("颜色主题", func() {
					// 切换主题
					if g.app.settings.Theme == "dark" {
						g.app.SetTheme("light")
					} else {
						g.app.SetTheme("dark")
					}
				}),
			),
			g.app.mainWindow.Canvas(),
		)
		popup.Show()
	})
	settingsBtn.Importance = widget.LowImportance

	// 设置按钮大小 - 使用更小的按钮
	buttons := []*widget.Button{explorerBtn, searchBtn, sourceControlBtn, runDebugBtn, extensionsBtn, settingsBtn}
	for _, btn := range buttons {
		btn.Resize(fyne.NewSize(ActivityBarWidth, ActivityBarWidth-8)) // 减小按钮高度
	}

	// 创建活动栏布局 - 更紧凑的间距
	topButtons := container.NewVBox(
		explorerBtn,
		searchBtn,
		sourceControlBtn,
		runDebugBtn,
		extensionsBtn,
	)

	// 去除内边距使布局更紧凑
	topButtons.Resize(fyne.NewSize(ActivityBarWidth, ActivityBarWidth*5-10))

	bottomButtons := container.NewVBox(
		settingsBtn,
	)

	// 使用Border容器将按钮分为顶部和底部
	return container.NewBorder(
		nil,           // top
		bottomButtons, // bottom
		nil, nil,      // left, right
		topButtons, // center - 移除内边距使布局更紧凑
	)
}

// createSideBar 创建侧边栏
func (g *MainGUI) createSideBar() *fyne.Container {
	// 侧边栏容器
	sidebar := container.NewVBox()

	// 设置固定大小
	sidebar.Resize(fyne.NewSize(g.sideBarWidth, g.app.mainWindow.Canvas().Size().Height))

	return sidebar
}

// createEditorArea 创建编辑器区域
func (g *MainGUI) createEditorArea() *fyne.Container {
	// 创建可关闭标签容器
	closableTabs := NewClosableTabs()

	// 添加第一个标签页
	closableTabs.AddTab("Untitled-1", g.editor.GetContent())

	// 设置标签关闭事件回调，确保至少有一个标签页
	closableTabs.OnTabClosed = func(index int) {
		// 如果没有标签了，创建一个新的
		if closableTabs.TabCount() == 0 {
			go func() {
				// 使用goroutine确保在UI更新完成后再添加新标签
				g.createNewFile()
			}()
		}
	}

	// 创建双击检测区域
	tabsContainer := NewDoubleTapContainer(closableTabs.GetContent(), func() {
		// 双击创建新标签页
		g.createNewFile()
	})

	// 使用 Max 容器来确保编辑器区域填充所有可用空间
	editorContainer := container.NewMax(tabsContainer)

	// 计算初始尺寸 - 确保足够大以填充可用空间
	if g.app != nil && g.app.mainWindow != nil {
		windowWidth := g.app.mainWindow.Canvas().Size().Width
		windowHeight := g.app.mainWindow.Canvas().Size().Height

		// 计算编辑器宽度
		editorWidth := windowWidth - ActivityBarWidth
		if g.sideBarVisible {
			editorWidth -= g.sideBarWidth
		}

		// 调整编辑器容器大小
		editorContainer.Resize(fyne.NewSize(editorWidth, windowHeight-StatusBarHeight))
	}

	return editorContainer
}

// DoubleTapContainer 是一个可双击的容器
type DoubleTapContainer struct {
	widget.BaseWidget
	Content     fyne.CanvasObject
	OnDoubleTap func()
}

// NewDoubleTapContainer 创建一个新的可双击容器
func NewDoubleTapContainer(content fyne.CanvasObject, onDoubleTap func()) *DoubleTapContainer {
	container := &DoubleTapContainer{
		Content:     content,
		OnDoubleTap: onDoubleTap,
	}
	container.ExtendBaseWidget(container)
	return container
}

// CreateRenderer 创建渲染器
func (d *DoubleTapContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.Content)
}

// DoubleTapped 处理双击事件
func (d *DoubleTapContainer) DoubleTapped(e *fyne.PointEvent) {
	if d.OnDoubleTap != nil {
		d.OnDoubleTap()
	}
}

// MinSize 返回最小尺寸
func (d *DoubleTapContainer) MinSize() fyne.Size {
	return d.Content.MinSize()
}

// Resize 调整大小
func (d *DoubleTapContainer) Resize(size fyne.Size) {
	d.BaseWidget.Resize(size)
	d.Content.Resize(size)
}

// Move 移动位置
func (d *DoubleTapContainer) Move(pos fyne.Position) {
	d.BaseWidget.Move(pos)
	d.Content.Move(pos)
}

// Show 显示容器
func (d *DoubleTapContainer) Show() {
	d.BaseWidget.Show()
	d.Content.Show()
}

// Hide 隐藏容器
func (d *DoubleTapContainer) Hide() {
	d.BaseWidget.Hide()
	d.Content.Hide()
}

// createStatusBar 创建状态栏
func (g *MainGUI) createStatusBar() *fyne.Container {
	// 创建状态栏内容 - 使用更小的字体
	left := widget.NewLabel("UTF-8")

	// 添加底部面板显示/隐藏按钮 - 使用合适的图标替代 TerminalIcon
	toggleTerminalBtn := widget.NewButtonWithIcon("", theme.ComputerIcon(), func() {
		g.toggleBottomPanel()
	})
	toggleTerminalBtn.Importance = widget.LowImportance
	// 调整按钮大小更小
	toggleTerminalBtn.Resize(fyne.NewSize(18, 18))

	right := container.NewHBox(
		toggleTerminalBtn,
		widget.NewLabel("Ln 1, Col 1"),
	)

	// 创建状态栏布局，确保其高度一致
	statusBar := container.NewBorder(
		nil, nil,
		left, right,
		nil,
	)

	// 设置固定高度
	statusBar.Resize(fyne.NewSize(1000, StatusBarHeight))

	return statusBar
}

// openFile 打开文件
func (g *MainGUI) openFile(path string) {
	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// 获取文件名
	fileName := filepath.Base(path)

	// 获取标签页容器
	var findClosableTabs func(obj fyne.CanvasObject) *ClosableTabs
	findClosableTabs = func(obj fyne.CanvasObject) *ClosableTabs {
		// 检查当前对象是否是DoubleTapContainer
		if dtc, ok := obj.(*DoubleTapContainer); ok {
			return findClosableTabs(dtc.Content)
		}

		// 检查当前对象是否是ClosableTabs的渲染内容
		if content, ok := obj.(*fyne.Container); ok {
			// 尝试从所有子组件中查找
			for _, child := range content.Objects {
				if tabs := findClosableTabs(child); tabs != nil {
					return tabs
				}
			}
		}

		// 直接检查是否是ClosableTabs
		if ct, ok := obj.(*ClosableTabs); ok {
			return ct
		}

		return nil
	}

	tabs := findClosableTabs(g.editorArea.Objects[0])

	if tabs != nil {
		// 创建新标签页并设置内容
		newEditor := NewEditorView()
		newEditor.SetText(string(content))

		// 添加新标签页
		tabIndex := tabs.AppendTab(fileName, newEditor.GetContent())

		// 选择新标签页
		tabs.SelectTab(tabIndex)
	} else {
		// 如果没有找到标签页容器，直接在当前编辑器中打开
		g.editor.SetText(string(content))
	}
}

// createNewFile 创建新文件
func (g *MainGUI) createNewFile() {
	// 获取编辑器区域中的容器
	maxContainer := g.editorArea.Objects[0]

	// 递归查找ClosableTabs组件
	var findClosableTabs func(obj fyne.CanvasObject) *ClosableTabs
	findClosableTabs = func(obj fyne.CanvasObject) *ClosableTabs {
		// 检查当前对象是否是DoubleTapContainer
		if dtc, ok := obj.(*DoubleTapContainer); ok {
			return findClosableTabs(dtc.Content)
		}

		// 检查当前对象是否是ClosableTabs的渲染内容
		if content, ok := obj.(*fyne.Container); ok {
			// 尝试从所有子组件中查找
			for _, child := range content.Objects {
				if tabs := findClosableTabs(child); tabs != nil {
					return tabs
				}
			}
		}

		// 直接检查是否是ClosableTabs
		if ct, ok := obj.(*ClosableTabs); ok {
			return ct
		}

		return nil
	}

	// 查找并获取标签页容器
	tabs := findClosableTabs(maxContainer)

	// 如果递归查找失败，可能容器在第一层
	if tabs == nil {
		// 尝试遍历所有顶层组件
		for _, obj := range g.editorArea.Objects {
			if foundTabs := findClosableTabs(obj); foundTabs != nil {
				tabs = foundTabs
				break
			}
		}
	}

	// 确保找到了标签页容器
	if tabs != nil {
		// 创建新的编辑器视图
		editorView := NewEditorView()

		// 计算新标签的序号（确保唯一性）
		tabCount := tabs.TabCount()
		newTabNumber := tabCount + 1

		// 创建标签页文本
		tabText := fmt.Sprintf("Untitled-%d", newTabNumber)

		// 添加新标签页
		tabs.AddTab(tabText, editorView.GetContent())

		// 选择新标签页
		tabs.SelectTab(tabs.TabCount() - 1)
	}
}

// openFolder 打开文件夹
func (g *MainGUI) openFolder() {
	// 这里应该打开一个文件夹选择对话框
	// 然后更新文件浏览器
	var popup *widget.PopUp
	content := container.NewVBox(
		widget.NewLabel("此处应显示文件夹选择对话框"),
		widget.NewButton("关闭", func() {
			if popup != nil {
				popup.Hide()
			}
		}),
	)
	popup = widget.NewModalPopUp(content, g.app.mainWindow.Canvas())
	popup.Show()
}

// refreshFileBrowser 刷新文件浏览器
func (g *MainGUI) refreshFileBrowser() {
	// 刷新文件浏览器内容
	g.updateSideBarContent()
}

// toggleBottomPanel 切换底部面板的显示状态
func (g *MainGUI) toggleBottomPanel() {
	// 切换底部面板显示状态
	g.bottomPanel.ToggleVisibility()

	// 更新布局
	g.updateLayout()

	// 刷新内容
	if g.app != nil && g.app.mainWindow != nil {
		g.app.mainWindow.SetContent(g.content)
		g.app.mainWindow.Canvas().Refresh(g.content)
	}
}
