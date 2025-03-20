package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ResizeableSidebar 可调整大小的侧边栏
type ResizeableSidebar struct {
	widget.BaseWidget
	content      fyne.CanvasObject
	handle       *resizeHandle
	currentWidth float32
	minWidth     float32
	maxWidth     float32
	onResize     func(float32)
}

// NewResizeableSidebar 创建一个新的可调整大小的侧边栏
func NewResizeableSidebar(content fyne.CanvasObject, width, minWidth, maxWidth float32, onResize func(float32)) *ResizeableSidebar {
	sidebar := &ResizeableSidebar{
		content:      content,
		currentWidth: width,
		minWidth:     minWidth,
		maxWidth:     maxWidth,
		onResize:     onResize,
	}

	sidebar.handle = newResizeHandle(sidebar)
	sidebar.ExtendBaseWidget(sidebar)

	return sidebar
}

// CreateRenderer 创建渲染器
func (s *ResizeableSidebar) CreateRenderer() fyne.WidgetRenderer {
	// 将调整手柄放在右侧
	border := container.NewBorder(nil, nil, nil, s.handle, s.content)
	return widget.NewSimpleRenderer(border)
}

// Resize 调整大小
func (s *ResizeableSidebar) Resize(size fyne.Size) {
	s.BaseWidget.Resize(size)
	s.currentWidth = size.Width
}

// MinSize 最小大小
func (s *ResizeableSidebar) MinSize() fyne.Size {
	return fyne.NewSize(s.minWidth, s.content.MinSize().Height)
}

// adjustWidth 调整宽度
func (s *ResizeableSidebar) adjustWidth(delta float32) {
	// 计算新的宽度
	newWidth := s.currentWidth + delta

	// 确保宽度在限制范围内
	if newWidth < s.minWidth {
		newWidth = s.minWidth
	} else if newWidth > s.maxWidth {
		newWidth = s.maxWidth
	}

	// 只有当宽度实际变化时才更新
	if newWidth != s.currentWidth {
		s.currentWidth = newWidth

		// 调用回调函数
		if s.onResize != nil {
			s.onResize(newWidth)
		}
	}
}

// resizeHandle 调整大小的手柄
type resizeHandle struct {
	widget.BaseWidget
	sidebar    *ResizeableSidebar
	isDragging bool
	startX     float32
	lastX      float32 // 记录上次拖动位置
	startWidth float32
	vertical   *canvas.Rectangle
}

// newResizeHandle 创建一个新的调整大小的手柄
func newResizeHandle(sidebar *ResizeableSidebar) *resizeHandle {
	handle := &resizeHandle{
		sidebar:  sidebar,
		vertical: canvas.NewRectangle(theme.PrimaryColor()), // 使用主题主色
	}

	handle.ExtendBaseWidget(handle)
	return handle
}

// CreateRenderer 创建渲染器
func (h *resizeHandle) CreateRenderer() fyne.WidgetRenderer {
	// 使用更宽的手柄，确保用户可以容易点击到
	h.vertical.Resize(fyne.NewSize(8, 1000))

	// 使用主题主色以更好地与主题匹配
	h.vertical.FillColor = theme.HoverColor()

	objects := []fyne.CanvasObject{h.vertical}
	return &resizeHandleRenderer{
		handle:  h,
		objects: objects,
	}
}

// resizeHandleRenderer 是resizeHandle的自定义渲染器
type resizeHandleRenderer struct {
	handle  *resizeHandle
	objects []fyne.CanvasObject
}

func (r *resizeHandleRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *resizeHandleRenderer) Destroy() {}

func (r *resizeHandleRenderer) Layout(size fyne.Size) {
	// 确保垂直条填满整个高度，并居中
	r.handle.vertical.Resize(fyne.NewSize(8, size.Height))
	r.handle.vertical.Move(fyne.NewPos(0, 0))
}

func (r *resizeHandleRenderer) MinSize() fyne.Size {
	return fyne.NewSize(8, 10)
}

func (r *resizeHandleRenderer) Refresh() {
	canvas.Refresh(r.handle.vertical)
}

// MinSize 最小大小
func (h *resizeHandle) MinSize() fyne.Size {
	return fyne.NewSize(8, 10)
}

// MouseDown 鼠标按下
func (h *resizeHandle) MouseDown(e *desktop.MouseEvent) {
	h.isDragging = true
	h.startX = e.Position.X
	h.lastX = e.Position.X
	h.startWidth = h.sidebar.currentWidth
	// 更改颜色以显示正在拖动
	h.vertical.FillColor = theme.FocusColor()
	canvas.Refresh(h.vertical)
	h.Refresh()
}

// MouseUp 鼠标抬起
func (h *resizeHandle) MouseUp(e *desktop.MouseEvent) {
	h.isDragging = false
	// 恢复颜色
	h.vertical.FillColor = theme.HoverColor()
	canvas.Refresh(h.vertical)
	h.Refresh()
}

// MouseMoved 鼠标移动
func (h *resizeHandle) MouseMoved(e *desktop.MouseEvent) {
	if h.isDragging {
		// 计算移动距离
		deltaX := e.Position.X - h.lastX

		// 更新最后位置
		h.lastX = e.Position.X

		// 调整宽度
		h.sidebar.adjustWidth(deltaX)

		// 立即刷新显示
		h.Refresh()
	}
}

// MouseIn 鼠标进入
func (h *resizeHandle) MouseIn(e *desktop.MouseEvent) {
	// 使调整手柄在鼠标悬停时更明显
	h.vertical.FillColor = theme.FocusColor()
	canvas.Refresh(h.vertical)
	h.Refresh()
}

// MouseOut 鼠标离开
func (h *resizeHandle) MouseOut() {
	// 如果不在拖动状态，则恢复正常颜色
	if !h.isDragging {
		h.vertical.FillColor = theme.HoverColor()
		canvas.Refresh(h.vertical)
		h.Refresh()
	}
}

// Cursor 光标
func (h *resizeHandle) Cursor() desktop.Cursor {
	return desktop.HResizeCursor
}
