package common

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type HasPrompt interface {
	GetPrompt() string
}

type HasImage interface {
	HasImage() bool
}

// openaiCompatiblePaths 定义需要 /v1 前缀的 OpenAI 兼容 API 路径
var openaiCompatiblePaths = []string{
	"/chat/completions",
	"/completions",
	"/embeddings",
	"/models",
	"/images/generations",
	"/images/edits",
	"/images/variations",
	"/audio/transcriptions",
	"/audio/translations",
	"/audio/speech",
	"/moderations",
	"/files",
	"/fine-tuning/jobs",
	"/batches",
	"/realtime",
	"/responses",
	"/rerank",
	"/assistants",
	"/threads",
	"/messages",
	"/runs",
	"/vector_stores",
}

// channelsRequiringV1Prefix 定义需要 /v1 前缀的渠道类型
// 这些渠道使用标准 OpenAI 兼容 API 格式
var channelsRequiringV1Prefix = map[int]bool{
	constant.ChannelTypeOpenAI:         true,
	constant.ChannelTypeOpenAIMax:      true,
	constant.ChannelTypeOhMyGPT:        true,
	constant.ChannelTypeAILS:           true,
	constant.ChannelTypeAIProxy:        true,
	constant.ChannelTypeAPI2GPT:        true,
	constant.ChannelTypeAIGC2D:         true,
	constant.ChannelTypeOpenRouter:     true,
	constant.ChannelTypeAIProxyLibrary: true,
	constant.ChannelTypeFastGPT:        true,
	constant.ChannelTypeMoonshot:       true,
	constant.ChannelTypePerplexity:     true,
	constant.ChannelTypeLingYiWanWu:    true,
	constant.ChannelTypeSiliconFlow:    true,
	constant.ChannelTypeMistral:        true,
	constant.ChannelTypeDeepSeek:       true,
	constant.ChannelTypeMokaAI:         true,
	constant.ChannelTypeXinference:     true,
	constant.ChannelTypeXai:            true,
	constant.ChannelTypeSubmodel:       true,
	constant.ChannelTypeSora:           true,
	constant.ChannelTypeOllama:         true,
	// 防呆设计：Gemini 渠道也可能使用 OpenAI 兼容反代（如 CatieCli），
	// 当通过 OpenAI 格式请求时需要 /v1 前缀
	constant.ChannelTypeGemini: true,
}

// NormalizeRequestPath 规范化请求路径，确保 OpenAI 兼容渠道有正确的 /v1 前缀
func NormalizeRequestPath(requestURL string, channelType int) string {
	// 如果渠道不需要 /v1 前缀，直接返回原路径
	if !channelsRequiringV1Prefix[channelType] {
		return requestURL
	}

	// 如果路径已经以 /v1 开头，直接返回
	if strings.HasPrefix(requestURL, "/v1/") || requestURL == "/v1" {
		return requestURL
	}

	// 检查是否是 OpenAI 兼容的 API 路径
	for _, path := range openaiCompatiblePaths {
		if strings.HasPrefix(requestURL, path) {
			// 添加 /v1 前缀
			return "/v1" + requestURL
		}
	}

	// 对于其他路径，保持不变
	return requestURL
}

func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	// 规范化请求路径：为需要 /v1 前缀的渠道自动添加前缀
	normalizedURL := NormalizeRequestPath(requestURL, channelType)
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, normalizedURL)

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		switch channelType {
		case constant.ChannelTypeOpenAI:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(normalizedURL, "/v1"))
		case constant.ChannelTypeAzure:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(normalizedURL, "/openai/deployments"))
		}
	}
	return fullRequestURL
}

func GetAPIVersion(c *gin.Context) string {
	query := c.Request.URL.Query()
	apiVersion := query.Get("api-version")
	if apiVersion == "" {
		apiVersion = c.GetString("api_version")
	}
	return apiVersion
}

func createTaskError(err error, code string, statusCode int, localError bool) *dto.TaskError {
	return &dto.TaskError{
		Code:       code,
		Message:    err.Error(),
		StatusCode: statusCode,
		LocalError: localError,
		Error:      err,
	}
}

func storeTaskRequest(c *gin.Context, info *RelayInfo, action string, requestObj TaskSubmitReq) {
	info.Action = action
	c.Set("task_request", requestObj)
}
func GetTaskRequest(c *gin.Context) (TaskSubmitReq, error) {
	v, exists := c.Get("task_request")
	if !exists {
		return TaskSubmitReq{}, fmt.Errorf("request not found in context")
	}
	req, ok := v.(TaskSubmitReq)
	if !ok {
		return TaskSubmitReq{}, fmt.Errorf("invalid task request type")
	}
	return req, nil
}

func validatePrompt(prompt string) *dto.TaskError {
	if strings.TrimSpace(prompt) == "" {
		return createTaskError(fmt.Errorf("prompt is required"), "invalid_request", http.StatusBadRequest, true)
	}
	return nil
}

func validateMultipartTaskRequest(c *gin.Context, info *RelayInfo, action string) (TaskSubmitReq, error) {
	var req TaskSubmitReq
	if _, err := c.MultipartForm(); err != nil {
		return req, err
	}

	formData := c.Request.PostForm
	req = TaskSubmitReq{
		Prompt:   formData.Get("prompt"),
		Model:    formData.Get("model"),
		Mode:     formData.Get("mode"),
		Image:    formData.Get("image"),
		Size:     formData.Get("size"),
		Metadata: make(map[string]interface{}),
	}

	if durationStr := formData.Get("seconds"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			req.Duration = duration
		}
	}

	if images := formData["images"]; len(images) > 0 {
		req.Images = images
	}

	for key, values := range formData {
		if len(values) > 0 && !isKnownTaskField(key) {
			if intVal, err := strconv.Atoi(values[0]); err == nil {
				req.Metadata[key] = intVal
			} else if floatVal, err := strconv.ParseFloat(values[0], 64); err == nil {
				req.Metadata[key] = floatVal
			} else {
				req.Metadata[key] = values[0]
			}
		}
	}
	return req, nil
}

func ValidateMultipartDirect(c *gin.Context, info *RelayInfo) *dto.TaskError {
	var prompt string
	var model string
	var seconds int
	var size string
	var hasInputReference bool

	var req TaskSubmitReq
	if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return createTaskError(err, "invalid_json", http.StatusBadRequest, true)
	}

	prompt = req.Prompt
	model = req.Model
	size = req.Size
	seconds, _ = strconv.Atoi(req.Seconds)
	if seconds == 0 {
		seconds = req.Duration
	}
	if req.InputReference != "" {
		req.Images = []string{req.InputReference}
	}

	if strings.TrimSpace(req.Model) == "" {
		return createTaskError(fmt.Errorf("model field is required"), "missing_model", http.StatusBadRequest, true)
	}

	if req.HasImage() {
		hasInputReference = true
	}

	if taskErr := validatePrompt(prompt); taskErr != nil {
		return taskErr
	}

	action := constant.TaskActionTextGenerate
	if hasInputReference {
		action = constant.TaskActionGenerate
	}
	if strings.HasPrefix(model, "sora-2") {

		if size == "" {
			size = "720x1280"
		}

		if seconds <= 0 {
			seconds = 4
		}

		if model == "sora-2" && !lo.Contains([]string{"720x1280", "1280x720"}, size) {
			return createTaskError(fmt.Errorf("sora-2 size is invalid"), "invalid_size", http.StatusBadRequest, true)
		}
		if model == "sora-2-pro" && !lo.Contains([]string{"720x1280", "1280x720", "1792x1024", "1024x1792"}, size) {
			return createTaskError(fmt.Errorf("sora-2 size is invalid"), "invalid_size", http.StatusBadRequest, true)
		}
		info.PriceData.OtherRatios = map[string]float64{
			"seconds": float64(seconds),
			"size":    1,
		}
		if lo.Contains([]string{"1792x1024", "1024x1792"}, size) {
			info.PriceData.OtherRatios["size"] = 1.666667
		}
	}

	info.Action = action

	return nil
}

func isKnownTaskField(field string) bool {
	knownFields := map[string]bool{
		"prompt":          true,
		"model":           true,
		"mode":            true,
		"image":           true,
		"images":          true,
		"size":            true,
		"duration":        true,
		"input_reference": true, // Sora 特有字段
	}
	return knownFields[field]
}

func ValidateBasicTaskRequest(c *gin.Context, info *RelayInfo, action string) *dto.TaskError {
	var err error
	contentType := c.GetHeader("Content-Type")
	var req TaskSubmitReq
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req, err = validateMultipartTaskRequest(c, info, action)
		if err != nil {
			return createTaskError(err, "invalid_multipart_form", http.StatusBadRequest, true)
		}
	} else if err := common.UnmarshalBodyReusable(c, &req); err != nil {
		return createTaskError(err, "invalid_request", http.StatusBadRequest, true)
	}

	if taskErr := validatePrompt(req.Prompt); taskErr != nil {
		return taskErr
	}

	if len(req.Images) == 0 && strings.TrimSpace(req.Image) != "" {
		// 兼容单图上传
		req.Images = []string{req.Image}
	}

	storeTaskRequest(c, info, action, req)
	return nil
}
