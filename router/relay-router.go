package router

import (
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/relay"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func SetRelayRouter(router *gin.Engine) {
	// 防呆设计：路径规范化已在 HTTP 层处理（main.go 中的 PathNormalizeHandler）
	// 处理双斜杠等问题，例如：//chat/completions -> /chat/completions
	router.Use(middleware.CORS())
	router.Use(middleware.DecompressRequestMiddleware())
	router.Use(middleware.StatsMiddleware())

	// ============================================================
	// 防呆设计：同时注册带 /v1 和不带 /v1 的路由
	// 这样无论用户配置 https://api.xxx.me 还是 https://api.xxx.me/v1
	// 都能正常工作
	// ============================================================

	// https://platform.openai.com/docs/api-reference/introduction
	// 注册 /v1/models 和 /models 两个路由组
	registerModelsRouter(router.Group("/v1/models"))
	registerModelsRouter(router.Group("/models")) // 防呆：不带 /v1 前缀

	// ============================================================
	// Google AI Studio (Gemini) 路由
	// 防呆设计：同时支持 /v1beta/... 和 /v1/v1beta/... 两种路径
	// ============================================================
	registerGeminiModelsRouter(router.Group("/v1beta/models"))
	registerGeminiModelsRouter(router.Group("/v1/v1beta/models")) // 防呆：支持 /v1 前缀

	registerGeminiCompatibleRouter(router.Group("/v1beta/openai/models"))
	registerGeminiCompatibleRouter(router.Group("/v1/v1beta/openai/models")) // 防呆：支持 /v1 前缀

	playgroundRouter := router.Group("/pg")
	playgroundRouter.Use(middleware.UserAuth(), middleware.Distribute())
	{
		playgroundRouter.POST("/chat/completions", controller.Playground)
	}

	// 注册 /v1 和根路径两个路由组的 API
	registerOpenAICompatibleRoutes(router.Group("/v1"))
	registerOpenAICompatibleRoutes(router.Group("")) // 防呆：不带 /v1 前缀

	relayMjRouter := router.Group("/mj")
	registerMjRouterGroup(relayMjRouter)

	relayMjModeRouter := router.Group("/:mode/mj")
	registerMjRouterGroup(relayMjModeRouter)
	//relayMjRouter.Use()

	relaySunoRouter := router.Group("/suno")
	relaySunoRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relaySunoRouter.POST("/submit/:action", controller.RelayTask)
		relaySunoRouter.POST("/fetch", controller.RelayTask)
		relaySunoRouter.GET("/fetch/:id", controller.RelayTask)
	}

	// Gemini API 路由
	// 防呆设计：同时支持 /v1beta/... 和 /v1/v1beta/... 两种路径
	registerGeminiRelayRouter(router.Group("/v1beta"))
	registerGeminiRelayRouter(router.Group("/v1/v1beta")) // 防呆：支持 /v1 前缀
}

// registerGeminiModelsRouter 注册 Gemini 模型列表路由
func registerGeminiModelsRouter(group *gin.RouterGroup) {
	group.Use(middleware.TokenAuth())
	{
		group.GET("", func(c *gin.Context) {
			controller.ListModels(c, constant.ChannelTypeGemini)
		})
	}
}

// registerGeminiCompatibleRouter 注册 Gemini 兼容的 OpenAI 模型列表路由
func registerGeminiCompatibleRouter(group *gin.RouterGroup) {
	group.Use(middleware.TokenAuth())
	{
		group.GET("", func(c *gin.Context) {
			controller.ListModels(c, constant.ChannelTypeOpenAI)
		})
	}
}

// registerGeminiRelayRouter 注册 Gemini API 中继路由
func registerGeminiRelayRouter(group *gin.RouterGroup) {
	group.Use(middleware.TokenAuth())
	group.Use(middleware.ModelRequestRateLimit())
	group.Use(middleware.Distribute())
	{
		// Gemini API 路径格式: /v1beta/models/{model_name}:{action}
		group.POST("/models/*path", func(c *gin.Context) {
			controller.Relay(c, types.RelayFormatGemini)
		})
	}
}

// registerModelsRouter 注册模型相关路由
// 用于同时支持 /v1/models 和 /models 两种路径
func registerModelsRouter(group *gin.RouterGroup) {
	group.Use(middleware.TokenAuth())
	{
		group.GET("", func(c *gin.Context) {
			switch {
			case c.GetHeader("x-api-key") != "" && c.GetHeader("anthropic-version") != "":
				controller.ListModels(c, constant.ChannelTypeAnthropic)
			case c.GetHeader("x-goog-api-key") != "" || c.Query("key") != "": // 单独的适配
				controller.RetrieveModel(c, constant.ChannelTypeGemini)
			default:
				controller.ListModels(c, constant.ChannelTypeOpenAI)
			}
		})

		group.GET("/:model", func(c *gin.Context) {
			switch {
			case c.GetHeader("x-api-key") != "" && c.GetHeader("anthropic-version") != "":
				controller.RetrieveModel(c, constant.ChannelTypeAnthropic)
			default:
				controller.RetrieveModel(c, constant.ChannelTypeOpenAI)
			}
		})
	}
}

// registerOpenAICompatibleRoutes 注册 OpenAI 兼容的 API 路由
// 用于同时支持 /v1/... 和 /... 两种路径（防呆设计）
func registerOpenAICompatibleRoutes(group *gin.RouterGroup) {
	group.Use(middleware.TokenAuth())
	group.Use(middleware.ModelRequestRateLimit())

	// WebSocket 路由（统一到 Relay）
	wsRouter := group.Group("")
	wsRouter.Use(middleware.Distribute())
	wsRouter.GET("/realtime", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIRealtime)
	})

	// HTTP 路由
	httpRouter := group.Group("")
	httpRouter.Use(middleware.Distribute())

	// claude related routes
	httpRouter.POST("/messages", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatClaude)
	})

	// chat related routes
	httpRouter.POST("/completions", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAI)
	})
	httpRouter.POST("/chat/completions", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAI)
	})

	// response related routes
	httpRouter.POST("/responses", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIResponses)
	})

	// image related routes
	httpRouter.POST("/edits", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIImage)
	})
	httpRouter.POST("/images/generations", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIImage)
	})
	httpRouter.POST("/images/edits", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIImage)
	})

	// embedding related routes
	httpRouter.POST("/embeddings", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatEmbedding)
	})

	// audio related routes
	httpRouter.POST("/audio/transcriptions", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIAudio)
	})
	httpRouter.POST("/audio/translations", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIAudio)
	})
	httpRouter.POST("/audio/speech", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAIAudio)
	})

	// rerank related routes
	httpRouter.POST("/rerank", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatRerank)
	})

	// gemini relay routes
	httpRouter.POST("/engines/:model/embeddings", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatGemini)
	})
	httpRouter.POST("/models/*path", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatGemini)
	})

	// other relay routes
	httpRouter.POST("/moderations", func(c *gin.Context) {
		controller.Relay(c, types.RelayFormatOpenAI)
	})

	// not implemented
	httpRouter.POST("/images/variations", controller.RelayNotImplemented)
	httpRouter.GET("/files", controller.RelayNotImplemented)
	httpRouter.POST("/files", controller.RelayNotImplemented)
	httpRouter.DELETE("/files/:id", controller.RelayNotImplemented)
	httpRouter.GET("/files/:id", controller.RelayNotImplemented)
	httpRouter.GET("/files/:id/content", controller.RelayNotImplemented)
	httpRouter.POST("/fine-tunes", controller.RelayNotImplemented)
	httpRouter.GET("/fine-tunes", controller.RelayNotImplemented)
	httpRouter.GET("/fine-tunes/:id", controller.RelayNotImplemented)
	httpRouter.POST("/fine-tunes/:id/cancel", controller.RelayNotImplemented)
	httpRouter.GET("/fine-tunes/:id/events", controller.RelayNotImplemented)
	httpRouter.DELETE("/models/:model", controller.RelayNotImplemented)
}

func registerMjRouterGroup(relayMjRouter *gin.RouterGroup) {
	relayMjRouter.GET("/image/:id", relay.RelayMidjourneyImage)
	relayMjRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		relayMjRouter.POST("/submit/action", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/shorten", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/modal", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/imagine", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/simple-change", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/describe", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/blend", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/edits", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/video", controller.RelayMidjourney)
		relayMjRouter.POST("/notify", controller.RelayMidjourney)
		relayMjRouter.GET("/task/:id/fetch", controller.RelayMidjourney)
		relayMjRouter.GET("/task/:id/image-seed", controller.RelayMidjourney)
		relayMjRouter.POST("/task/list-by-condition", controller.RelayMidjourney)
		relayMjRouter.POST("/insight-face/swap", controller.RelayMidjourney)
		relayMjRouter.POST("/submit/upload-discord-images", controller.RelayMidjourney)
	}
}
