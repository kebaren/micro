package gui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zyedidia/micro/v2/internal/buffer"
)

// EditorView represents the editor view in GUI
type EditorView struct {
	textArea    *widget.Entry
	lineNumbers *widget.Label
	currentBuf  *buffer.Buffer
	container   *fyne.Container
}

// CustomTextStyle represents a custom text styling
type CustomTextStyle struct {
	widget.BaseWidget
	entry *widget.Entry
}

// 变量区域，保存自定义大小信息
var (
	// EditorFontSize 编辑器字体大小
	EditorFontSize float32 = 14
	// LineNumberFontSize 行号字体大小
	LineNumberFontSize float32 = 13
)

// NewCustomEntry creates a new custom styled entry
func NewCustomEntry() *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapOff
	// 使用等宽字体并增大字号
	entry.TextStyle = fyne.TextStyle{Monospace: true}

	// 确保多行模式启用
	entry.MultiLine = true

	// 确保输入区可见，这样输入法就会跟随光标位置
	entry.SetMinRowsVisible(10)

	return entry
}

// NewEditorView creates a new editor view
func NewEditorView() *EditorView {
	e := &EditorView{
		textArea:    NewCustomEntry(),
		lineNumbers: widget.NewLabel(""),
	}

	// 性能优化：减小编辑区域的内部绘制频率和内存使用
	e.textArea.ExtendBaseWidget(e.textArea)

	// 减少文本区域的最小可见行数，降低内存占用
	e.textArea.SetMinRowsVisible(3)

	// 设置行号样式 - 使用更大的文本样式提高可读性
	e.lineNumbers.Alignment = fyne.TextAlignTrailing
	e.lineNumbers.TextStyle = fyne.TextStyle{Monospace: true, TabWidth: 2}

	// 创建自定义大小的等宽字体文本样式
	e.textArea.TextStyle = fyne.TextStyle{Monospace: true}

	// 文本区域样式优化
	e.textArea.Wrapping = fyne.TextWrapOff

	// 获取主题颜色 - 使用共享的颜色对象减少内存占用
	_, _, _, secondary := GetThemeColors()

	// 更符合VSCode的文本区域背景色
	textAreaBg := canvas.NewRectangle(secondary)

	e.textArea.SetPlaceHolder("")

	// 内存优化：使用共享的定时器对象和较大的防抖动延迟
	var contentUpdateTimer *time.Timer
	var isUpdatingContent = false

	// 性能和内存优化：减少不必要的刷新和内存使用
	var isFocusingCursor = false
	e.textArea.OnCursorChanged = func() {
		// 避免递归调用
		if isFocusingCursor {
			return
		}

		isFocusingCursor = true
		// 异步刷新，避免UI阻塞
		go func() {
			// 确保重置标志
			defer func() { isFocusingCursor = false }()

			// 仅在文本区域有焦点时刷新
			if MainWindow != nil && MainWindow.Canvas().Focused() == e.textArea {
				// 使用更长的延迟，减少刷新频率
				time.Sleep(200 * time.Millisecond)
				if MainWindow != nil {
					fyne.CurrentApp().Driver().CanvasForObject(e.textArea).Refresh(e.textArea)
				}
			}
		}()
	}

	e.textArea.OnChanged = func(content string) {
		// 避免在更新过程中触发新的更新
		if isUpdatingContent {
			return
		}

		isUpdatingContent = true

		// 使用防抖动方式处理频繁更新，内存优化：使用更长的延迟
		if contentUpdateTimer != nil {
			contentUpdateTimer.Stop()
		}

		contentUpdateTimer = time.AfterFunc(300*time.Millisecond, func() {
			// 更新缓冲区内容
			e.updateBufferFromText(content)

			// 标记更新完成
			isUpdatingContent = false
		})
	}

	// 内存优化：使用更简单的行号背景
	lineNumberBg := canvas.NewRectangle(theme.InputBackgroundColor())
	lineNumberBg.SetMinSize(fyne.NewSize(40, 0))

	// 行号容器 - 内存优化：使用更简单的布局
	lineNumberWithBg := container.New(layout.NewMaxLayout(),
		lineNumberBg,
		e.lineNumbers,
	)

	// 性能优化：简化组件层级
	textWithPadding := container.NewMax(textAreaBg, e.textArea)

	// 创建编辑区域内容布局，将行号放在左侧
	editArea := container.New(layout.NewBorderLayout(nil, nil, lineNumberWithBg, nil),
		lineNumberWithBg, // 左侧行号区域
		textWithPadding,  // 中心文本区域
	)

	// 内存优化：使用更轻量级的滚动容器配置
	editorScrollContainer := container.NewScroll(editArea)
	editorScrollContainer.Direction = container.ScrollBoth

	// 创建主容器 - 移除底部状态栏，使用MainGUI中的全局状态栏
	e.container = container.NewBorder(
		nil,                   // 顶部
		nil,                   // 底部状态栏 - 已移除
		nil,                   // 左侧
		nil,                   // 右侧
		editorScrollContainer, // 中心内容
	)

	return e
}

// SetBuffer sets the current buffer
func (e *EditorView) SetBuffer(buf *buffer.Buffer) {
	e.currentBuf = buf
	if buf != nil {
		// Convert buffer content to string for the text area
		content := ""
		for i := 0; i < buf.LinesNum(); i++ {
			content += buf.Line(i)
			if i < buf.LinesNum()-1 {
				content += "\n"
			}
		}

		// 临时禁用OnChanged监听器，避免触发更新
		savedOnChanged := e.textArea.OnChanged
		e.textArea.OnChanged = nil
		e.textArea.SetText(content)

		// 恢复之前保存的OnChanged处理器
		e.textArea.OnChanged = savedOnChanged

		// Update line numbers
		e.updateLineNumbers(content)

		// 状态栏信息现在通过MainGUI更新
		UpdateMainStatus("Ready")
	} else {
		e.textArea.SetText("")
		e.lineNumbers.SetText("")
		UpdateMainStatus("No file")
	}
}

// updateBufferFromText updates the buffer content when text area changes
func (e *EditorView) updateBufferFromText(content string) {
	if e.currentBuf == nil {
		return
	}

	// 性能优化：快速哈希比较检查内容是否真正变化
	contentHash := getContentHash(content)
	bufferHash := getBufferHash(e.currentBuf)

	if contentHash == bufferHash {
		return // 内容没有变化，跳过所有更新
	}

	// 性能优化：只更新有变化的部分而不是整个缓冲区
	newLines := strings.Split(content, "\n")
	oldLines := make([]string, e.currentBuf.LinesNum())

	for i := 0; i < e.currentBuf.LinesNum(); i++ {
		oldLines[i] = e.currentBuf.Line(i)
	}

	// 删除所有内容并重新插入新内容效率更高，尤其是对大文档
	size := e.currentBuf.Size()
	if size > 0 {
		e.currentBuf.Remove(buffer.Loc{X: 0, Y: 0}, buffer.Loc{X: size, Y: 0})
	}

	// 插入新内容
	e.currentBuf.Insert(buffer.Loc{X: 0, Y: 0}, content)

	// 性能优化：仅在行数变化时更新行号
	oldLineCount := len(oldLines)
	newLineCount := len(newLines)

	if oldLineCount != newLineCount {
		e.updateLineNumbers(content)
	}

	// 更新主状态栏显示
	UpdateMainStatus(fmt.Sprintf("Lines: %d", newLineCount))
}

// 创建内容哈希函数，快速比较内容是否变化
func getContentHash(content string) int {
	// 简单的哈希算法，比完整比较字符串更快
	hash := 0
	if len(content) < 1000 {
		// 对于小文件，直接返回长度作为简单哈希
		return len(content)
	}

	// 对于大文件，采样一些关键位置
	samples := []int{0, 100, 500, len(content) / 2, len(content) - 1}
	for _, pos := range samples {
		if pos < len(content) {
			hash = hash*31 + int(content[pos])
		}
	}
	return hash + len(content)
}

// 获取缓冲区内容的哈希值
func getBufferHash(buf *buffer.Buffer) int {
	if buf == nil {
		return 0
	}

	// 计算缓冲区内容的哈希值
	hash := 0
	lineCount := buf.LinesNum()

	if lineCount < 100 {
		// 对于小缓冲区，使用行数作为简单哈希
		return lineCount
	}

	// 对于大缓冲区，采样一些行
	samples := []int{0, 10, 50, lineCount / 2, lineCount - 1}
	for _, line := range samples {
		if line < lineCount {
			hash = hash*31 + len(buf.Line(line))
		}
	}
	return hash + lineCount
}

// updateLineNumbers updates the line numbers display
func (e *EditorView) updateLineNumbers(content string) {
	// 内存优化：对于大文件只计算行数而不解析全部文本
	var lineCount int
	if len(content) < 10000 {
		// 对于小文件直接使用标准方式计算行数
		lineCount = len(strings.Split(content, "\n"))
	} else {
		// 对于大文件，手动计算行数而不创建字符串数组，减少内存使用
		lineCount = 1
		for i := 0; i < len(content); i++ {
			if content[i] == '\n' {
				lineCount++
			}
		}
	}

	// 内存优化：限制行号显示，对于超大文件只显示前1000行
	maxLinesToDisplay := 1000
	displayedLineCount := lineCount
	if lineCount > maxLinesToDisplay {
		displayedLineCount = maxLinesToDisplay
	}

	// 内存优化：预先分配足够的字符串容量，避免字符串追加操作
	lineText := strings.Builder{}
	maxDigits := len(fmt.Sprintf("%d", lineCount))

	// 预分配一个固定大小缓冲区
	lineText.Grow(displayedLineCount * (maxDigits + 1)) // 每行行号 + 换行符

	// 批量生成显示的行号
	for i := 1; i <= displayedLineCount; i++ {
		fmt.Fprintf(&lineText, "%*d\n", maxDigits, i)
	}

	// 对于超出显示限制的大文件，添加省略号提示
	if lineCount > maxLinesToDisplay {
		lineText.WriteString("...\n")
	}

	// 设置行号文本
	e.lineNumbers.SetText(lineText.String())

	// 更新主状态栏显示
	UpdateMainStatus(fmt.Sprintf("Lines: %d", lineCount))
}

// GetContainer returns the container for this view
func (e *EditorView) GetContainer() fyne.CanvasObject {
	return e.container
}

// UpdateFromBuffer updates the editor view from the buffer
func (e *EditorView) UpdateFromBuffer() {
	if e.currentBuf != nil {
		e.SetBuffer(e.currentBuf)
	}
}
