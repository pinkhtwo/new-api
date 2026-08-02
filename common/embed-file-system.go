package common

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/static"
)

// Credit: https://github.com/gin-contrib/static/issues/19

// apiPathPrefixes 是不应该被静态文件服务处理的 API 路径前缀
// 防呆设计：确保 API 请求不会被静态文件服务拦截
var apiPathPrefixes = []string{
	"/v1",
	"/v1beta",
	"/api",
	"/chat",
	"/completions",
	"/embeddings",
	"/images",
	"/audio",
	"/models",
	"/messages",
	"/moderations",
	"/edits",
	"/files",
	"/fine-tunes",
	"/fine-tuning",
	"/responses",
	"/realtime",
	"/rerank",
	"/engines",
	"/mj",
	"/suno",
	"/pg",
}

// isAPIPath 检查给定路径是否是 API 路径
func isAPIPath(path string) bool {
	for _, prefix := range apiPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

type embedFileSystem struct {
	http.FileSystem
}

func (e *embedFileSystem) Exists(prefix string, path string) bool {
	// 防呆设计：API 路径不应该被静态文件服务处理
	// 即使 web/dist 目录中意外包含了这些路径，也跳过
	if isAPIPath(path) {
		return false
	}
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func (e *embedFileSystem) Open(name string) (http.File, error) {
	if name == "/" {
		// This will make sure the index page goes to NoRouter handler,
		// which will use the replaced index bytes with analytic codes.
		return nil, os.ErrNotExist
	}
	return e.FileSystem.Open(name)
}

func EmbedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	efs, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return &embedFileSystem{
		FileSystem: http.FS(efs),
	}
}
