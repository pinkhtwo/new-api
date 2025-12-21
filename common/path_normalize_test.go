package common

import (
	"testing"
)

// TestExtractAPIEndpoint 测试 API 端点智能提取功能
func TestExtractAPIEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// ============================================================
		// 正常路径（无需修改）
		// ============================================================
		{
			name:     "正常 OpenAI chat completions 路径",
			input:    "/v1/chat/completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "正常 Gemini 路径",
			input:    "/v1beta/models/gemini-pro:generateContent",
			expected: "/v1beta/models/gemini-pro:generateContent",
		},
		{
			name:     "正常 Claude 路径",
			input:    "/v1/messages",
			expected: "/v1/messages",
		},
		{
			name:     "正常模型列表路径",
			input:    "/v1/models",
			expected: "/v1/models",
		},

		// ============================================================
		// 用户错误添加前缀的情况（防呆设计核心场景）
		// ============================================================
		{
			name:     "错误前缀：/ABC/v1/chat/completions",
			input:    "/ABC/v1/chat/completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "中文前缀：/我是奶龙/v1/chat/completions",
			input:    "/我是奶龙/v1/chat/completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "多级错误前缀：/test/abc/v1/chat/completions",
			input:    "/test/abc/v1/chat/completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "错误前缀 Gemini：/ABC/v1beta/models/gemini-pro:generateContent",
			input:    "/ABC/v1beta/models/gemini-pro:generateContent",
			expected: "/v1beta/models/gemini-pro:generateContent",
		},
		{
			name:     "错误前缀 Claude：/test/v1/messages",
			input:    "/test/v1/messages",
			expected: "/v1/messages",
		},
		{
			name:     "错误前缀不带 v1：/ABC/chat/completions",
			input:    "/ABC/chat/completions",
			expected: "/chat/completions",
		},

		// ============================================================
		// Gemini 带 v1 前缀的特殊情况
		// ============================================================
		{
			name:     "Gemini 带 v1 前缀：/v1/v1beta/models/gemini-pro:generateContent",
			input:    "/v1/v1beta/models/gemini-pro:generateContent",
			expected: "/v1/v1beta/models/gemini-pro:generateContent",
		},
		{
			name:     "错误前缀 Gemini 带 v1：/ABC/v1/v1beta/models/gemini-pro:generateContent",
			input:    "/ABC/v1/v1beta/models/gemini-pro:generateContent",
			expected: "/v1/v1beta/models/gemini-pro:generateContent",
		},

		// ============================================================
		// 模型列表端点
		// ============================================================
		{
			name:     "错误前缀模型列表：/ABC/v1/models",
			input:    "/ABC/v1/models",
			expected: "/v1/models",
		},
		{
			name:     "错误前缀模型详情：/test/v1/models/gpt-4",
			input:    "/test/v1/models/gpt-4",
			expected: "/v1/models/gpt-4",
		},

		// ============================================================
		// 其他 API 端点
		// ============================================================
		{
			name:     "错误前缀 embeddings：/test/v1/embeddings",
			input:    "/test/v1/embeddings",
			expected: "/v1/embeddings",
		},
		{
			name:     "错误前缀 images：/ABC/v1/images/generations",
			input:    "/ABC/v1/images/generations",
			expected: "/v1/images/generations",
		},
		{
			name:     "错误前缀 audio：/test/v1/audio/speech",
			input:    "/test/v1/audio/speech",
			expected: "/v1/audio/speech",
		},
		{
			name:     "错误前缀 rerank：/ABC/v1/rerank",
			input:    "/ABC/v1/rerank",
			expected: "/v1/rerank",
		},

		// ============================================================
		// Midjourney 和其他特殊 API
		// ============================================================
		{
			name:     "错误前缀 Midjourney：/ABC/mj/submit/imagine",
			input:    "/ABC/mj/submit/imagine",
			expected: "/mj/submit/imagine",
		},
		{
			name:     "错误前缀 Suno：/test/suno/submit/music",
			input:    "/test/suno/submit/music",
			expected: "/suno/submit/music",
		},

		// ============================================================
		// Dashboard API
		// ============================================================
		{
			name:     "错误前缀 dashboard：/ABC/dashboard/billing/subscription",
			input:    "/ABC/dashboard/billing/subscription",
			expected: "/dashboard/billing/subscription",
		},
		{
			name:     "错误前缀 v1 dashboard：/test/v1/dashboard/billing/usage",
			input:    "/test/v1/dashboard/billing/usage",
			expected: "/dashboard/billing/usage", // 优先匹配 /dashboard/ 端点
		},

		// ============================================================
		// 管理 API
		// ============================================================
		{
			name:     "错误前缀 API：/ABC/api/status",
			input:    "/ABC/api/status",
			expected: "/api/status",
		},

		// ============================================================
		// 未知路径（应保持不变）
		// ============================================================
		{
			name:     "未知路径应保持不变",
			input:    "/unknown/path",
			expected: "/unknown/path",
		},
		{
			name:     "根路径",
			input:    "/",
			expected: "/",
		},
		{
			name:     "静态资源路径",
			input:    "/assets/js/app.js",
			expected: "/assets/js/app.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractAPIEndpoint(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractAPIEndpoint(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizePath 测试路径规范化功能
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "双斜杠规范化",
			input:    "//v1/chat/completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "多个斜杠规范化",
			input:    "///v1///chat//completions",
			expected: "/v1/chat/completions",
		},
		{
			name:     "保留尾部斜杠",
			input:    "/v1/models/",
			expected: "/v1/models/",
		},
		{
			name:     "根路径",
			input:    "/",
			expected: "/",
		},
		{
			name:     "正常路径不变",
			input:    "/v1/chat/completions",
			expected: "/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeAndExtractPath 集成测试：测试完整的路径处理流程
func TestNormalizeAndExtractPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedAPI string
	}{
		{
			name:        "用户输入 https://api.xxx/ABC 后 SillyTavern 拼接的 OpenAI 路径",
			input:       "/ABC/v1/chat/completions",
			expectedAPI: "/v1/chat/completions",
		},
		{
			name:        "用户输入 https://api.xxx/我是奶龙 后 SillyTavern 拼接的 Gemini 路径",
			input:       "/我是奶龙/v1beta/models/gemini-pro:generateContent",
			expectedAPI: "/v1beta/models/gemini-pro:generateContent",
		},
		{
			name:        "用户输入 https://api.xxx///// 后的双斜杠问题",
			input:       "//////v1/chat/completions",
			expectedAPI: "/v1/chat/completions",
		},
		{
			name:        "用户输入 https://api.xxx///v1 后的路径",
			input:       "///v1/v1/chat/completions",
			expectedAPI: "/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAndExtractPath(tt.input)
			if result != tt.expectedAPI {
				t.Errorf("NormalizeAndExtractPath(%q) = %q, want %q", tt.input, result, tt.expectedAPI)
			}
		})
	}
}
