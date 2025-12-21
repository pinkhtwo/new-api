package common

import (
	"path"
	"strings"
)

// apiEndpoints 定义所有需要识别的 API 端点
// 按优先级排序：更长/更具体的路径优先匹配
// 防呆设计：支持用户在 URL 中添加任意前缀后仍能正确路由
var apiEndpoints = []string{
	// ============================================================
	// Gemini API 端点（更长的路径优先）
	// 注意：同时包含带尾部斜杠和不带尾部斜杠的版本
	// ============================================================
	"/v1/v1beta/openai/models", // Gemini 兼容 OpenAI 模型列表（带 v1 前缀）
	"/v1beta/openai/models",    // Gemini 兼容 OpenAI 模型列表
	"/v1/v1beta/models/",       // Gemini API（带 v1 前缀，带尾部斜杠）
	"/v1/v1beta/models",        // Gemini API（带 v1 前缀，不带尾部斜杠）
	"/v1beta/models/",          // Gemini API（带尾部斜杠）
	"/v1beta/models",           // Gemini API（不带尾部斜杠）- 重要：用于 SillyTavern

	// ============================================================
	// OpenAI 兼容 API 端点
	// ============================================================
	"/v1/chat/completions",   // OpenAI Chat Completions
	"/chat/completions",      // OpenAI Chat Completions（不带 v1）
	"/v1/completions",        // OpenAI Completions
	"/completions",           // OpenAI Completions（不带 v1）
	"/v1/responses",          // OpenAI Responses
	"/responses",             // OpenAI Responses（不带 v1）
	"/v1/embeddings",         // OpenAI Embeddings
	"/embeddings",            // OpenAI Embeddings（不带 v1）
	"/v1/images/generations", // OpenAI Image Generation
	"/images/generations",    // OpenAI Image Generation（不带 v1）
	"/v1/images/edits",       // OpenAI Image Edits
	"/images/edits",          // OpenAI Image Edits（不带 v1）
	"/v1/audio/transcriptions",
	"/audio/transcriptions",
	"/v1/audio/translations",
	"/audio/translations",
	"/v1/audio/speech",
	"/audio/speech",
	"/v1/moderations",
	"/moderations",
	"/v1/edits",
	"/edits",
	"/v1/rerank",
	"/rerank",
	"/v1/realtime",
	"/realtime",

	// ============================================================
	// Claude API 端点
	// ============================================================
	"/v1/messages",
	"/messages",

	// ============================================================
	// 模型列表端点
	// ============================================================
	"/v1/models/",
	"/v1/models",
	"/models/",
	"/models",

	// ============================================================
	// 其他 API 端点（Midjourney、Suno 等）
	// ============================================================
	"/mj/",
	"/suno/",
	"/pg/",
	"/kling/",
	"/jimeng/",
	"/v1/video/",
	"/v1/videos/",

	// ============================================================
	// Dashboard API
	// ============================================================
	"/dashboard/",
	"/v1/dashboard/",

	// ============================================================
	// 管理 API
	// ============================================================
	"/api/",
}

// NormalizePath 规范化路径
// 1. 移除多余的斜杠 (// -> /)
// 2. 保持路径的结构
func NormalizePath(p string) string {
	// 使用 path.Clean 来规范化路径
	// 但是 path.Clean 会移除尾部斜杠，我们需要保留它
	hasTrailingSlash := len(p) > 1 && p[len(p)-1] == '/'

	// 规范化路径
	cleaned := path.Clean(p)

	// 确保路径以 / 开头
	if !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}

	// 如果原路径有尾部斜杠且不是根路径，保留它
	if hasTrailingSlash && cleaned != "/" {
		cleaned += "/"
	}

	return cleaned
}

// ExtractAPIEndpoint 从路径中智能提取 API 端点
// 防呆设计：无论用户在 URL 中添加什么前缀，都能正确识别并提取 API 端点
// 例如：
//   - /ABC/v1/chat/completions -> /v1/chat/completions
//   - /我是奶龙/v1beta/models/gemini-pro:generateContent -> /v1beta/models/gemini-pro:generateContent
//   - /test/v1/models -> /v1/models
func ExtractAPIEndpoint(p string) string {
	// 如果路径以 /api/ 开头，不进行提取（管理 API 不需要防呆处理）
	if strings.HasPrefix(p, "/api/") {
		return p
	}

	// 遍历所有已知的 API 端点
	for _, endpoint := range apiEndpoints {
		// 查找端点在路径中的位置
		idx := strings.Index(p, endpoint)
		if idx != -1 {
			// 找到了端点，提取从端点开始的完整路径
			return p[idx:]
		}
	}

	// 未找到已知端点，返回原始路径
	return p
}

// NormalizeAndExtractPath 规范化路径并提取 API 端点
// 这是一个便捷函数，组合了 NormalizePath 和 ExtractAPIEndpoint
func NormalizeAndExtractPath(p string) string {
	normalized := NormalizePath(p)
	return ExtractAPIEndpoint(normalized)
}
