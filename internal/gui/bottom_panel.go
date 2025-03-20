package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BottomPanel 表示底部面板
type BottomPanel struct {
	content  fyne.CanvasObject
	tabs     *container.AppTabs
	terminal *widget.Entry
	output   *widget.Entry
	problems *widget.Entry
	visible  bool // 添加可见性状态
}

// NewBottomPanel 创建新的底部面板
func NewBottomPanel() *BottomPanel {
	panel := &BottomPanel{
		visible: true, // 默认可见
	}

	// 创建终端 - 使用更紧凑的布局
	panel.terminal = widget.NewMultiLineEntry()
	panel.terminal.Resize(fyne.NewSize(800, 150)) // 设置较小的默认高度

	// 创建输出
	panel.output = widget.NewMultiLineEntry()
	panel.output.Resize(fyne.NewSize(800, 150))

	// 创建问题
	panel.problems = widget.NewMultiLineEntry()
	panel.problems.Resize(fyne.NewSize(800, 150))

	// 创建标签页 - 使用更小的标签
	panel.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.ComputerIcon(), panel.terminal),
		container.NewTabItemWithIcon("", theme.DocumentIcon(), panel.output),
		container.NewTabItemWithIcon("", theme.WarningIcon(), panel.problems),
	)

	// 设置小标签位置
	panel.tabs.SetTabLocation(container.TabLocationBottom)

	// 创建一个固定高度的容器
	fixedHeightContainer := container.NewVBox(
		panel.tabs,
	)
	fixedHeightContainer.Resize(fyne.NewSize(800, 160)) // 设置固定高度

	// 创建底部面板布局
	panel.content = container.NewMax(fixedHeightContainer)
	panel.content.Resize(fyne.NewSize(800, 160)) // 设置固定高度

	return panel
}

// GetContent 获取底部面板内容
func (b *BottomPanel) GetContent() fyne.CanvasObject {
	return b.content
}

// ShowTerminal 显示终端
func (b *BottomPanel) ShowTerminal() {
	b.tabs.SelectIndex(0)
}

// ShowOutput 显示输出
func (b *BottomPanel) ShowOutput() {
	b.tabs.SelectIndex(1)
}

// ShowProblems 显示问题
func (b *BottomPanel) ShowProblems() {
	b.tabs.SelectIndex(2)
}

// WriteToTerminal 写入终端
func (b *BottomPanel) WriteToTerminal(text string) {
	b.terminal.SetText(text)
}

// WriteToOutput 写入输出
func (b *BottomPanel) WriteToOutput(text string) {
	b.output.SetText(text)
}

// WriteToProblems 写入问题
func (b *BottomPanel) WriteToProblems(text string) {
	b.problems.SetText(text)
}

// IsVisible 返回面板是否可见
func (b *BottomPanel) IsVisible() bool {
	return b.visible
}

// ToggleVisibility 切换面板可见性
func (b *BottomPanel) ToggleVisibility() {
	b.visible = !b.visible
}

// Show 显示面板
func (b *BottomPanel) Show() {
	b.visible = true
}

// Hide 隐藏面板
func (b *BottomPanel) Hide() {
	b.visible = false
}
