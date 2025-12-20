package router

import (
	"embed"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// apiPrefixes 是所有 API 路径前缀
// 防呆设计：同时支持带 /v1 和不带 /v1 的 API 路径
var apiPrefixes = []string{
	"/v1",     // OpenAI 兼容 API
	"/v1beta", // Gemini API
	"/api",    // 管理 API
	"/assets", // 静态资源
	"/mj",     // Midjourney
	"/suno",   // Suno
	"/pg",     // Playground
	// 防呆设计：不带 /v1 前缀的 OpenAI 兼容 API 路径
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
}

// isAPIRequest 检查请求 URI 是否是 API 请求
// 防呆设计：路径规范化已在 HTTP 层处理（main.go 中的 PathNormalizeHandler）
func isAPIRequest(uri string) bool {
	for _, prefix := range apiPrefixes {
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}
	return false
}

func SetWebRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte) {
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", common.EmbedFolder(buildFS, "web/dist")))
	router.NoRoute(func(c *gin.Context) {
		// 检查是否是 API 请求路径
		// 防呆设计：支持带 /v1 和不带 /v1 的 API 路径
		// 路径规范化已在 HTTP 层处理（main.go 中的 PathNormalizeHandler）
		if isAPIRequest(c.Request.URL.Path) {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPage)
	})
}
