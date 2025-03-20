package gui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FileBrowser 表示文件浏览器
type FileBrowser struct {
	content    fyne.CanvasObject
	tree       *widget.Tree
	rootPath   string
	onFileOpen func(string)
}

// NewFileBrowser 创建新的文件浏览器
func NewFileBrowser(rootPath string, onFileOpen func(string)) *FileBrowser {
	browser := &FileBrowser{
		rootPath:   rootPath,
		onFileOpen: onFileOpen,
	}

	// 创建文件树
	browser.tree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return browser.getChildPaths(uid)
		},
		func(uid widget.TreeNodeID) bool {
			return browser.isDirectory(uid)
		},
		func(isBranch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(uid widget.TreeNodeID, isBranch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			if uid == "" {
				label.SetText("/")
			} else {
				fullPath := filepath.Join(browser.rootPath, uid)
				info, err := os.Stat(fullPath)
				if err == nil {
					label.SetText(info.Name())
				} else {
					label.SetText(uid)
				}
			}
		},
	)

	// 创建文件浏览器布局
	browser.content = container.NewBorder(
		widget.NewLabel("EXPLORER"),
		nil,
		nil,
		nil,
		browser.tree,
	)

	return browser
}

// GetContent 获取文件浏览器内容
func (f *FileBrowser) GetContent() fyne.CanvasObject {
	return f.content
}

// getChildPaths 获取子路径
func (f *FileBrowser) getChildPaths(path string) []string {
	if path == "" {
		return []string{"/"}
	}

	fullPath := filepath.Join(f.rootPath, path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			paths = append(paths, filepath.Join(path, entry.Name()))
		}
	}
	return paths
}

// isDirectory 判断是否为目录
func (f *FileBrowser) isDirectory(path string) bool {
	if path == "" {
		return true
	}

	fullPath := filepath.Join(f.rootPath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
