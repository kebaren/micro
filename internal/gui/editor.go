package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// EditorView 表示编辑器视图
type EditorView struct {
	content     fyne.CanvasObject
	lineNumbers *widget.TextGrid
	editor      *widget.Entry
	minimap     *widget.TextGrid
}

// NewEditorView 创建新的编辑器视图
func NewEditorView() *EditorView {
	view := &EditorView{}

	// 创建行号 - 使用更窄的宽度
	view.lineNumbers = widget.NewTextGrid()
	view.lineNumbers.SetText("1\n2\n3\n4\n5")
	// 设置较窄的行号宽度
	view.lineNumbers.Resize(fyne.NewSize(24, 1000))
	lineNumBg := canvas.NewRectangle(theme.BackgroundColor())
	lineNumContainer := container.NewMax(lineNumBg, view.lineNumbers)

	// 创建编辑器
	view.editor = widget.NewMultiLineEntry()
	view.editor.SetPlaceHolder("Enter text here...")
	// 调整编辑器样式
	view.editor.MultiLine = true
	view.editor.Wrapping = fyne.TextWrapWord

	// 自定义编辑器字体和尺寸
	if theme.TextSize() > 14 {
		// 如果主题字体大小过大，则调整为适中大小
		// 注意：这只是示意，Fyne实际不支持直接调整单个组件的字体大小
	}

	// 设置最小尺寸以确保编辑器区域足够大
	view.editor.Resize(fyne.NewSize(800, 600))

	// 添加边框和背景
	editorBg := canvas.NewRectangle(theme.BackgroundColor())
	editorBorder := canvas.NewRectangle(theme.ShadowColor())
	editorBorder.StrokeWidth = 1
	editorBorder.StrokeColor = theme.ShadowColor()
	editorContainer := container.NewMax(editorBg, editorBorder, view.editor)

	// 创建小地图（缩略视图）- 使用更窄的宽度
	view.minimap = widget.NewTextGrid()
	view.minimap.SetText("")
	// 设置更窄的小地图宽度
	view.minimap.Resize(fyne.NewSize(50, 1000))
	minimapBg := canvas.NewRectangle(theme.BackgroundColor())
	minimapContainer := container.NewMax(minimapBg, view.minimap)

	// 创建编辑器布局 - 使用 Border 布局确保编辑器填充所有可用空间
	view.content = container.NewBorder(
		nil, nil, // 上下无内容
		lineNumContainer, // 左侧行号
		minimapContainer, // 右侧小地图
		editorContainer,  // 中间编辑器填满可用空间
	)

	return view
}

// GetContent 获取编辑器视图内容
func (e *EditorView) GetContent() fyne.CanvasObject {
	return e.content
}

// SetText 设置编辑器文本
func (e *EditorView) SetText(text string) {
	e.editor.SetText(text)
	// 更新行号
	e.updateLineNumbers(text)
	// 更新小地图
	e.updateMinimap(text)
}

// GetText 获取编辑器文本
func (e *EditorView) GetText() string {
	return e.editor.Text
}

// updateLineNumbers 更新行号
func (e *EditorView) updateLineNumbers(text string) {
	lineCount := strings.Count(text, "\n") + 1
	lineNumbers := make([]string, lineCount)
	for i := 0; i < lineCount; i++ {
		// 使用更紧凑的行号格式
		lineNumbers[i] = fmt.Sprintf("%d", i+1)
	}
	e.lineNumbers.SetText(strings.Join(lineNumbers, "\n"))
}

// updateMinimap 更新小地图
func (e *EditorView) updateMinimap(text string) {
	// 简化的小地图视图
	lines := strings.Split(text, "\n")
	minimapText := ""
	for _, line := range lines {
		if len(line) > 0 {
			minimapText += "■"
		} else {
			minimapText += "□"
		}
		minimapText += "\n"
	}
	e.minimap.SetText(minimapText)
}
