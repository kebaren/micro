package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ClosableTabItem 表示可关闭的标签项
type ClosableTabItem struct {
	Title   string
	Content fyne.CanvasObject
	OnClose func()
	TabItem *container.TabItem
}

// ClosableTabs 表示带有关闭按钮的标签容器
type ClosableTabs struct {
	widget.BaseWidget
	container     *fyne.Container
	tabBar        *fyne.Container
	contentArea   *fyne.Container
	items         []*ClosableTabItem
	selectedIndex int
	OnTabClosed   func(index int)
}

// NewClosableTabs 创建新的可关闭标签容器
func NewClosableTabs() *ClosableTabs {
	ct := &ClosableTabs{
		tabBar:        container.NewHBox(),
		contentArea:   container.NewMax(),
		items:         []*ClosableTabItem{},
		selectedIndex: -1,
	}

	// 创建一个更紧凑的主容器
	ct.container = container.NewBorder(
		container.NewVBox(
			ct.tabBar,
			widget.NewSeparator(), // 添加分隔线更清晰
		), nil, nil, nil,
		ct.contentArea,
	)

	// 限制tab栏高度
	ct.tabBar.Resize(fyne.NewSize(1000, 28))

	ct.ExtendBaseWidget(ct)
	return ct
}

// CreateRenderer 创建渲染器
func (ct *ClosableTabs) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ct.container)
}

// AddTab 添加标签页
func (ct *ClosableTabs) AddTab(title string, content fyne.CanvasObject) *ClosableTabItem {
	// 创建标签按钮和关闭按钮 - 使用更小的字号和紧凑布局
	tabButton := widget.NewButton(title, nil)
	// 设置为低重要性让按钮更加扁平
	tabButton.Importance = widget.LowImportance
	// 设置小尺寸的按钮
	tabButton.Resize(fyne.NewSize(100, 28))

	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), nil)
	closeButton.Importance = widget.LowImportance
	// 设置更小的关闭按钮尺寸
	closeButton.Resize(fyne.NewSize(16, 16))

	// 创建更紧凑的标签栏项
	tabItem := container.NewBorder(
		nil, nil, nil,
		closeButton,
		tabButton,
	)
	// 设置较小的内边距
	tabItem.Resize(fyne.NewSize(120, 28))

	// 创建TabItem
	item := &ClosableTabItem{
		Title:   title,
		Content: content,
	}

	// 保存到items列表
	ct.items = append(ct.items, item)
	index := len(ct.items) - 1

	// 设置标签按钮点击事件
	tabButton.OnTapped = func() {
		ct.SelectTab(index)
	}

	// 设置关闭按钮点击事件
	closeButton.OnTapped = func() {
		ct.CloseTab(index)
	}

	// 将标签栏项添加到标签栏
	ct.tabBar.Add(tabItem)

	// 如果这是第一个标签，则选中它
	if len(ct.items) == 1 {
		ct.SelectTab(0)
	}

	return item
}

// AppendTab 添加标签页并返回索引
func (ct *ClosableTabs) AppendTab(title string, content fyne.CanvasObject) int {
	ct.AddTab(title, content)
	return len(ct.items) - 1
}

// CloseTab 关闭标签页
func (ct *ClosableTabs) CloseTab(index int) {
	if index < 0 || index >= len(ct.items) {
		return
	}

	// 获取要关闭的标签项
	closeItem := ct.items[index]

	// 如果有自定义关闭处理函数，调用它
	if closeItem.OnClose != nil {
		closeItem.OnClose()
	}

	// 从tabBar中移除标签按钮
	ct.tabBar.Remove(ct.tabBar.Objects[index])

	// 从items列表中移除标签项
	ct.items = append(ct.items[:index], ct.items[index+1:]...)

	// 调整选中的标签索引
	if len(ct.items) > 0 {
		if ct.selectedIndex == index {
			// 如果关闭的是当前选中的标签
			if index == len(ct.items) {
				// 如果关闭的是最后一个标签，选择前一个
				ct.SelectTab(index - 1)
			} else {
				// 否则选择同一位置的标签（现在是新的标签）
				ct.SelectTab(index)
			}
		} else if ct.selectedIndex > index {
			// 如果关闭的标签在当前选中标签之前，需要调整索引
			ct.selectedIndex--
		}
	} else {
		// 如果没有标签了，清空内容区域
		ct.contentArea.Objects = []fyne.CanvasObject{}
		ct.selectedIndex = -1
	}

	// 触发标签关闭事件
	if ct.OnTabClosed != nil {
		ct.OnTabClosed(index)
	}

	// 刷新界面
	ct.Refresh()
}

// SelectTab 选择标签页
func (ct *ClosableTabs) SelectTab(index int) {
	if index < 0 || index >= len(ct.items) {
		return
	}

	// 更新选中的标签索引
	ct.selectedIndex = index

	// 更新标签按钮样式
	for i := range ct.items {
		tabButton := ct.getTabButton(i)
		if i == index {
			tabButton.Importance = widget.HighImportance
		} else {
			tabButton.Importance = widget.MediumImportance
		}
	}

	// 更新内容区域
	ct.contentArea.Objects = []fyne.CanvasObject{ct.items[index].Content}

	// 刷新界面
	ct.Refresh()
}

// getTabButton 获取标签按钮
func (ct *ClosableTabs) getTabButton(index int) *widget.Button {
	if index < 0 || index >= len(ct.tabBar.Objects) {
		return nil
	}

	// 标签按钮包装在Border容器中
	border := ct.tabBar.Objects[index].(*fyne.Container)

	// 标签按钮在Border的中心位置
	// 通常情况下我们可以通过border.Objects获取中心对象，
	// 但由于container.NewBorder的实现，我们需要使用下面的方法：
	for _, obj := range border.Objects {
		if btn, ok := obj.(*widget.Button); ok {
			if closeBtn, ok := obj.(*widget.Button); !ok || closeBtn.Icon != theme.CancelIcon() {
				return btn
			}
		}
	}

	return nil
}

// SelectedIndex 获取当前选中的标签索引
func (ct *ClosableTabs) SelectedIndex() int {
	return ct.selectedIndex
}

// TabCount 获取标签数量
func (ct *ClosableTabs) TabCount() int {
	return len(ct.items)
}

// Refresh 刷新界面
func (ct *ClosableTabs) Refresh() {
	ct.tabBar.Refresh()
	ct.contentArea.Refresh()
	ct.container.Refresh()
}

// MinSize 返回最小尺寸
func (ct *ClosableTabs) MinSize() fyne.Size {
	return ct.container.MinSize()
}

// GetContent 获取内容
func (ct *ClosableTabs) GetContent() fyne.CanvasObject {
	return ct.container
}
