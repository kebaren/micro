package gui

import (
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zyedidia/micro/v2/internal/buffer"
)

// MainGUI represents the main GUI of the application
type MainGUI struct {
	editorViews        map[int]*EditorView
	content            *fyne.Container
	currentBufferIndex int
	tabs               *container.DocTabs
	toolbar            *widget.Toolbar
	contentContainer   *fyne.Container
	sidePanel          *fyne.Container
	fileBrowser        *fyne.Container
	isFilePanelVisible bool
	splitContainer     *container.Split
}

// NewMainGUI creates a new main GUI
func NewMainGUI() *MainGUI {
	m := &MainGUI{
		editorViews:        make(map[int]*EditorView),
		currentBufferIndex: -1,
		isFilePanelVisible: false,
	}

	// 创建全局状态栏 - 作为单一状态显示区域
	MainStatus = widget.NewLabel("Ready")
	MainStatus.TextStyle = fyne.TextStyle{Monospace: true, Bold: false}
	MainStatus.Alignment = fyne.TextAlignLeading

	// 获取主题颜色
	bg, _, _, _ := GetThemeColors()

	// 创建标签页 - 性能优化：延迟初始化内容
	m.tabs = container.NewDocTabs()
	m.tabs.SetTabLocation(container.TabLocationTop)

	// 创建主要内容区域容器 - 使用MaxLayout可以减少重绘操作
	contentContainer := container.NewMax()

	// 性能优化：限制标签页变更后的重绘频率
	var tabChangeTimer *time.Timer
	m.tabs.OnSelected = func(tab *container.TabItem) {
		// 内存优化：如果有上一个计时器，停止它并清理内存
		if tabChangeTimer != nil {
			tabChangeTimer.Stop()
			tabChangeTimer = nil
		}

		// 使用更轻量的延迟处理，避免频繁切换标签时的内存占用
		tabChangeTimer = time.AfterFunc(100*time.Millisecond, func() {
			// 查找对应的缓冲区索引
			for i, buf := range buffer.OpenBuffers {
				if tab.Text == filepath.Base(buf.GetName()) {
					// 仅在索引变化时更新
					if m.currentBufferIndex != i {
						m.currentBufferIndex = i

						// 更新内容容器显示当前选中的编辑器
						if edView, exists := m.editorViews[i]; exists && edView != nil {
							// 内存优化：直接替换而非添加删除对象
							if len(contentContainer.Objects) > 0 {
								contentContainer.Objects[0] = edView.GetContainer()
							} else {
								contentContainer.Objects = []fyne.CanvasObject{edView.GetContainer()}
							}
							contentContainer.Refresh()
						}
					}
					break
				}
			}

			// 清理计时器引用
			tabChangeTimer = nil
		})
	}

	// 为双击检测设置变量
	var lastClickTime int64
	// 处理Tab空白区域点击 - 实现双击检测但减少频繁创建
	m.tabs.OnUnselected = func(tab *container.TabItem) {
		// 检查是否在空白区域点击
		if tab == nil {
			currentTime := time.Now().UnixMilli()
			if currentTime-lastClickTime < 500 { // 500ms内的两次点击视为双击
				// 延迟创建新文件，避免UI冻结
				go func() {
					time.Sleep(10 * time.Millisecond)
					m.CreateNewTab()
				}()
			}
			lastClickTime = currentTime
		}
	}

	// VSCode风格的新建标签按钮 - 使用主题颜色
	newTabButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		// 创建新文件
		m.CreateNewTab()
	})
	newTabButton.Importance = widget.LowImportance
	newTabButton.Resize(fyne.NewSize(24, 24))

	// 创建文件浏览器面板
	m.createFileBrowser()

	// VSCode风格的活动栏布局 - 性能优化：降低活动栏复杂度
	activityBar := container.NewVBox()

	// 背景矩形使整个左侧面板更加明显
	activityBarBg := canvas.NewRectangle(bg)
	activityBarBg.SetMinSize(fyne.NewSize(48, 0))

	// 创建侧边栏按钮 - 添加文件浏览器功能
	fileExplorerButton := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		// 切换文件浏览器面板的可见性
		m.toggleFileBrowser()
	})
	fileExplorerButton.Importance = widget.LowImportance

	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		// 搜索功能，暂不实现
		UpdateMainStatus("Search function not implemented yet")
	})
	searchButton.Importance = widget.LowImportance

	// 添加按钮到活动栏
	activityBar.Add(fileExplorerButton)
	activityBar.Add(searchButton)

	// 添加设置按钮到活动栏底部
	settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		// 设置功能暂不实现
		UpdateMainStatus("Settings panel not implemented yet")
	})
	settingsButton.Importance = widget.LowImportance

	// 添加一个弹性空间，将设置按钮推到底部
	activityBar.Add(layout.NewSpacer())
	activityBar.Add(settingsButton)

	// 创建左侧面板容器，确保活动栏有固定宽度
	activityBarContainer := container.New(layout.NewMaxLayout(), activityBarBg, activityBar)

	// 创建侧边面板容器，默认包含文件浏览器但不显示
	m.sidePanel = container.NewMax()

	// 设置所有标签右侧的按钮
	tabButtons := container.NewHBox(newTabButton)

	// 创建VSCode风格布局 - 性能优化：减少容器层级
	var topBarContainer fyne.CanvasObject

	if AppSettings.ShowToolbar {
		m.toolbar = m.createToolbar()

		// 创建顶部区域：工具栏和标签栏
		topBarContainer = container.NewBorder(
			m.toolbar,
			nil,
			nil,
			tabButtons,
			m.tabs,
		)
	} else {
		// 不显示工具栏，创建更简洁的VSCode风格布局
		topBarContainer = container.NewBorder(
			nil,
			nil,
			nil,
			tabButtons,
			m.tabs,
		)
	}

	// 准备侧边内容容器 - 活动栏和可展开区域
	sidePanelWithActivityBar := container.New(
		layout.NewBorderLayout(nil, nil, activityBarContainer, nil),
		activityBarContainer,
		m.sidePanel,
	)

	// 使用Split容器实现可调整宽度的侧边栏
	// 默认设置侧边栏宽度为0（不可见），可以通过拖动调整宽度
	m.splitContainer = container.NewHSplit(
		sidePanelWithActivityBar,
		contentContainer,
	)

	// 设置初始分割比例，0表示完全隐藏左侧面板，只显示活动栏
	m.splitContainer.SetOffset(0)

	// 主布局 - 使用全局状态栏替代本地状态栏
	m.content = container.NewBorder(
		topBarContainer,  // 顶部区域 - 包含标签页
		MainStatus,       // 底部状态栏 - 使用全局状态栏
		nil,              // 左侧为空，因为已经在splitContent中设置
		nil,              // 右侧
		m.splitContainer, // 中心区域是可调整大小的分割内容
	)

	// 保存内容容器的引用，方便后续访问
	m.contentContainer = contentContainer

	// 设置菜单栏，包含工具栏功能按钮
	m.setupMainMenu()

	return m
}

// setupMainMenu creates the main menu for the application with toolbar functions
func (m *MainGUI) setupMainMenu() {
	// 文件菜单
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() { m.newFile() }),
		fyne.NewMenuItem("Open", func() { m.openFile() }),
		fyne.NewMenuItem("Save", func() { m.saveFile() }),
		fyne.NewMenuItem("Save As", func() { m.saveFileAs() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { App.Quit() }),
	)

	// 编辑菜单
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutCut{Clipboard: MainWindow.Clipboard()})
			}
		}),
		fyne.NewMenuItem("Copy", func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutCopy{Clipboard: MainWindow.Clipboard()})
			}
		}),
		fyne.NewMenuItem("Paste", func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutPaste{Clipboard: MainWindow.Clipboard()})
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Find", func() {
			// 未实现
		}),
		fyne.NewMenuItem("Replace", func() {
			// 未实现
		}),
	)

	// 视图菜单
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Show Toolbar", func() {
			AppSettings.ShowToolbar = !AppSettings.ShowToolbar
			// 重新创建UI
			newGUI := NewMainGUI()
			MainWindow.SetContent(newGUI.GetContent())
			MainWindow.Show()
		}),
		fyne.NewMenuItem("Toggle Theme", func() {
			AppSettings.UseDarkTheme = !AppSettings.UseDarkTheme
			App.Settings().SetTheme(newCompactTheme())
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Syntax Highlighting", func() {}),
		fyne.NewMenuItem("Line Numbers", func() {}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Full Screen", func() {
			MainWindow.SetFullScreen(!MainWindow.FullScreen())
		}),
	)

	// 使用更大字体的主菜单
	mainMenu := fyne.NewMainMenu(
		fileMenu,
		editMenu,
		viewMenu,
	)

	MainWindow.SetMainMenu(mainMenu)
}

// createToolbar creates a toolbar with common actions
func (m *MainGUI) createToolbar() *widget.Toolbar {
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() { m.newFile() }),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() { m.openFile() }),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() { m.saveFile() }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutCut{Clipboard: MainWindow.Clipboard()})
			}
		}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutCopy{Clipboard: MainWindow.Clipboard()})
			}
		}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			if edView := m.getCurrentEditorView(); edView != nil {
				edView.textArea.TypedShortcut(&fyne.ShortcutPaste{Clipboard: MainWindow.Clipboard()})
			}
		}),
	)

	return toolbar
}

// getCurrentEditorView returns the currently active editor view
func (m *MainGUI) getCurrentEditorView() *EditorView {
	if m.currentBufferIndex >= 0 && m.currentBufferIndex < len(buffer.OpenBuffers) {
		return m.editorViews[m.currentBufferIndex]
	}
	return nil
}

// addTab adds a new tab for the given buffer
func (m *MainGUI) addTab(buf *buffer.Buffer) {
	// 创建新的编辑器视图
	editorView := NewEditorView()

	// 设置缓冲区到编辑器视图
	editorView.SetBuffer(buf)

	// 保存编辑器视图以便后续访问
	bufferIndex := len(buffer.OpenBuffers) - 1
	m.editorViews[bufferIndex] = editorView

	// 获取文件名作为标签名
	tabName := "Untitled"
	if buf.Path != "" {
		tabName = filepath.Base(buf.Path)
	} else if buf.GetName() != "" {
		tabName = filepath.Base(buf.GetName())
	}

	// 创建空容器作为标签页内容 (实际内容会在选择时动态显示)
	emptyContainer := container.NewMax()
	tab := container.NewTabItem(tabName, emptyContainer)

	// 添加到标签页组并选中
	m.tabs.Append(tab)
	m.tabs.Select(tab)

	// 直接更新主内容区域，避免通过标签选择触发
	m.currentBufferIndex = bufferIndex
	m.contentContainer.Objects = []fyne.CanvasObject{editorView.GetContainer()}

	// 优化：只刷新必要的组件
	m.contentContainer.Refresh()
}

// newFile creates a new file
func (m *MainGUI) newFile() {
	// 使用公开的CreateNewTab方法
	m.CreateNewTab()
}

// openFile opens a file dialog to open a file
func (m *MainGUI) openFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		buf, err := buffer.NewBufferFromFile(path, buffer.BTDefault)
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}

		// Check if file is already open
		for i, existingBuf := range buffer.OpenBuffers {
			if existingBuf.Path == path {
				// File already open, just switch to its tab
				m.currentBufferIndex = i
				for j, item := range m.tabs.Items {
					if item.Text == filepath.Base(path) {
						m.tabs.Select(m.tabs.Items[j])
						return
					}
				}
				return
			}
		}

		buffer.OpenBuffers = append(buffer.OpenBuffers, buf)
		m.currentBufferIndex = len(buffer.OpenBuffers) - 1
		m.addTab(buf)
	}, MainWindow)
}

// saveFile saves the current file
func (m *MainGUI) saveFile() {
	if m.currentBufferIndex < 0 || m.currentBufferIndex >= len(buffer.OpenBuffers) {
		return
	}

	buf := buffer.OpenBuffers[m.currentBufferIndex]
	if buf.Path == "" {
		m.saveFileAs()
		return
	}

	// The buffer should already be updated by the editor view's change listener
	err := buf.Save()
	if err != nil {
		dialog.ShowError(err, MainWindow)
	}

	// Update tab name
	for i, item := range m.tabs.Items {
		if i == m.tabs.SelectedIndex() {
			item.Text = filepath.Base(buf.GetName())
			m.tabs.Refresh()
			break
		}
	}
}

// saveFileAs opens a file dialog to save the current file with a new name
func (m *MainGUI) saveFileAs() {
	if m.currentBufferIndex < 0 || m.currentBufferIndex >= len(buffer.OpenBuffers) {
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		buf := buffer.OpenBuffers[m.currentBufferIndex]
		buf.Path = writer.URI().Path()

		// The buffer should already be updated by the editor view's change listener
		err = buf.Save()
		if err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}

		// Update tab name
		for i, item := range m.tabs.Items {
			if i == m.tabs.SelectedIndex() {
				item.Text = filepath.Base(buf.GetName())
				m.tabs.Refresh()
				break
			}
		}
	}, MainWindow)
}

// GetContent returns the main content container
func (m *MainGUI) GetContent() fyne.CanvasObject {
	return m.content
}

// CreateNewTab creates a new empty tab
func (m *MainGUI) CreateNewTab() {
	// 使用原有的newFile方法逻辑，创建一个新的空白标签页
	// 创建一个真正的空白缓冲区而不是复制现有内容
	buf := buffer.NewBufferFromString("", "Untitled", buffer.BTDefault)

	// 添加到全局缓冲区列表
	buffer.OpenBuffers = append(buffer.OpenBuffers, buf)

	// 设置当前缓冲区索引
	m.currentBufferIndex = len(buffer.OpenBuffers) - 1

	// 创建新的编辑器视图
	editorView := NewEditorView()

	// 直接设置缓冲区到编辑器视图，确保是一个干净的视图
	editorView.SetBuffer(buf)

	// 保存编辑器视图以便后续访问
	bufferIndex := len(buffer.OpenBuffers) - 1
	m.editorViews[bufferIndex] = editorView

	// 获取文件名作为标签名
	tabName := "Untitled"

	// 创建空容器作为标签页内容
	emptyContainer := container.NewMax()
	tab := container.NewTabItem(tabName, emptyContainer)

	// 添加到标签页组并选中
	m.tabs.Append(tab)
	m.tabs.Select(tab)

	// 直接更新主内容区域
	m.contentContainer.Objects = []fyne.CanvasObject{editorView.GetContainer()}

	// 仅刷新内容容器，避免刷新整个UI
	m.contentContainer.Refresh()

	// 设置状态栏信息 - 使用全局状态栏更新
	UpdateMainStatus("New file created")
}

// createFileBrowser 创建文件浏览器面板
func (m *MainGUI) createFileBrowser() {
	// 创建文件浏览器标题
	title := widget.NewLabel("文件浏览器")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 创建一个目录树视图
	list := widget.NewList(
		func() int { return 3 }, // 示例项目数量
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FileIcon()),
				widget.NewLabel("文件名"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			icon := box.Objects[0].(*widget.Icon)

			switch id {
			case 0:
				label.SetText("文件夹")
				icon.SetResource(theme.FolderIcon())
			case 1:
				label.SetText("README.md")
				icon.SetResource(theme.FileTextIcon())
			case 2:
				label.SetText("main.go")
				icon.SetResource(theme.FileIcon())
			}
		},
	)

	// 添加文件点击处理
	list.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			// 文件夹点击
			UpdateMainStatus("点击了文件夹")
		case 1:
			// README.md点击
			UpdateMainStatus("点击了 README.md")
		case 2:
			// main.go点击
			UpdateMainStatus("点击了 main.go")
		}
		// 取消选中状态
		list.UnselectAll()
	}

	// 创建搜索输入框
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("搜索文件...")

	// 创建工具按钮
	refreshButton := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		UpdateMainStatus("刷新文件列表")
	})
	newFolderButton := widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
		UpdateMainStatus("新建文件夹")
	})
	newFileButton := widget.NewButtonWithIcon("", theme.FileIcon(), func() {
		UpdateMainStatus("新建文件")
	})

	// 工具栏布局
	toolBar := container.NewHBox(
		refreshButton,
		newFolderButton,
		newFileButton,
	)

	// 完整文件浏览器面板
	content := container.NewBorder(
		container.NewVBox(
			title,
			searchEntry,
			toolBar,
		),
		nil,
		nil,
		nil,
		list,
	)

	// 设置背景色
	_, _, _, secondary := GetThemeColors()
	bg := canvas.NewRectangle(secondary)

	// 适应可变宽度的面板
	// 设置面板初始大小，但允许自由调整
	content.Resize(fyne.NewSize(220, 0))

	// 创建可滚动的容器，确保在宽度变化时内容仍然可以完整查看
	scrollContainer := container.NewScroll(content)
	scrollContainer.Direction = container.ScrollBoth

	// 保存文件浏览器引用，使用更灵活的容器
	m.fileBrowser = container.New(layout.NewMaxLayout(), bg, scrollContainer)
}

// toggleFileBrowser 切换文件浏览器面板的可见性
func (m *MainGUI) toggleFileBrowser() {
	m.isFilePanelVisible = !m.isFilePanelVisible

	if m.isFilePanelVisible {
		// 显示文件浏览器
		m.sidePanel.Objects = []fyne.CanvasObject{m.fileBrowser}
		m.sidePanel.Refresh()

		// 设置侧边栏宽度为默认值(0.2表示占总宽度的20%)
		m.splitContainer.SetOffset(0.2)

		UpdateMainStatus("File browser opened")
	} else {
		// 隐藏文件浏览器
		m.sidePanel.Objects = []fyne.CanvasObject{}
		m.sidePanel.Refresh()

		// 重置侧边栏宽度为最小值(0表示只显示活动栏)
		m.splitContainer.SetOffset(0)

		UpdateMainStatus("File browser closed")
	}
}
