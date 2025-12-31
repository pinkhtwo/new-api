package constant

import (
	"net/http"
	"strings"
)

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModerations
	RelayModeImagesGenerations
	RelayModeImagesEdits
	RelayModeEdits

	RelayModeMidjourneyImagine
	RelayModeMidjourneyDescribe
	RelayModeMidjourneyBlend
	RelayModeMidjourneyChange
	RelayModeMidjourneySimpleChange
	RelayModeMidjourneyNotify
	RelayModeMidjourneyTaskFetch
	RelayModeMidjourneyTaskImageSeed
	RelayModeMidjourneyTaskFetchByCondition
	RelayModeMidjourneyAction
	RelayModeMidjourneyModal
	RelayModeMidjourneyShorten
	RelayModeSwapFace
	RelayModeMidjourneyUpload
	RelayModeMidjourneyVideo
	RelayModeMidjourneyEdits

	RelayModeAudioSpeech        // tts
	RelayModeAudioTranscription // whisper
	RelayModeAudioTranslation   // whisper

	RelayModeSunoFetch
	RelayModeSunoFetchByID
	RelayModeSunoSubmit

	RelayModeVideoFetchByID
	RelayModeVideoSubmit

	RelayModeRerank

	RelayModeResponses

	RelayModeRealtime

	RelayModeGemini
)

func Path2RelayMode(path string) int {
	relayMode := RelayModeUnknown

	// ============================================================
	// 防呆设计：同时支持带 /v1 和不带 /v1 前缀的路径
	// 这样无论用户配置的 URL 是否带 /v1，都能正确识别 RelayMode
	// ============================================================

	// Chat Completions
	if strings.HasPrefix(path, "/v1/chat/completions") ||
		strings.HasPrefix(path, "/chat/completions") ||
		strings.HasPrefix(path, "/pg/chat/completions") {
		relayMode = RelayModeChatCompletions
	} else if strings.HasPrefix(path, "/v1/completions") || strings.HasPrefix(path, "/completions") {
		// Completions (legacy)
		relayMode = RelayModeCompletions
	} else if strings.HasPrefix(path, "/v1/embeddings") || strings.HasPrefix(path, "/embeddings") {
		// Embeddings
		relayMode = RelayModeEmbeddings
	} else if strings.HasSuffix(path, "embeddings") {
		// Embeddings (catch-all for various embedding endpoints)
		relayMode = RelayModeEmbeddings
	} else if strings.HasPrefix(path, "/v1/moderations") || strings.HasPrefix(path, "/moderations") {
		// Moderations
		relayMode = RelayModeModerations
	} else if strings.HasPrefix(path, "/v1/images/generations") || strings.HasPrefix(path, "/images/generations") {
		// Image Generations
		relayMode = RelayModeImagesGenerations
	} else if strings.HasPrefix(path, "/v1/images/edits") || strings.HasPrefix(path, "/images/edits") {
		// Image Edits
		relayMode = RelayModeImagesEdits
	} else if strings.HasPrefix(path, "/v1/edits") || strings.HasPrefix(path, "/edits") {
		// Edits (legacy)
		relayMode = RelayModeEdits
	} else if strings.HasPrefix(path, "/v1/responses") || strings.HasPrefix(path, "/responses") {
		// Responses
		relayMode = RelayModeResponses
	} else if strings.HasPrefix(path, "/v1/audio/speech") || strings.HasPrefix(path, "/audio/speech") {
		// Audio Speech (TTS)
		relayMode = RelayModeAudioSpeech
	} else if strings.HasPrefix(path, "/v1/audio/transcriptions") || strings.HasPrefix(path, "/audio/transcriptions") {
		// Audio Transcriptions (Whisper)
		relayMode = RelayModeAudioTranscription
	} else if strings.HasPrefix(path, "/v1/audio/translations") || strings.HasPrefix(path, "/audio/translations") {
		// Audio Translations (Whisper)
		relayMode = RelayModeAudioTranslation
	} else if strings.HasPrefix(path, "/v1/rerank") || strings.HasPrefix(path, "/rerank") {
		// Rerank
		relayMode = RelayModeRerank
	} else if strings.HasPrefix(path, "/v1/realtime") || strings.HasPrefix(path, "/realtime") {
		// Realtime (WebSocket)
		relayMode = RelayModeRealtime
	} else if strings.HasPrefix(path, "/v1beta/models") || strings.HasPrefix(path, "/v1/models") || strings.HasPrefix(path, "/models") {
		// Gemini models
		relayMode = RelayModeGemini
	} else if strings.HasPrefix(path, "/mj") {
		// Midjourney
		relayMode = Path2RelayModeMidjourney(path)
	}
	return relayMode
}

func Path2RelayModeMidjourney(path string) int {
	relayMode := RelayModeUnknown
	if strings.HasSuffix(path, "/mj/submit/action") {
		// midjourney plus
		relayMode = RelayModeMidjourneyAction
	} else if strings.HasSuffix(path, "/mj/submit/modal") {
		// midjourney plus
		relayMode = RelayModeMidjourneyModal
	} else if strings.HasSuffix(path, "/mj/submit/shorten") {
		// midjourney plus
		relayMode = RelayModeMidjourneyShorten
	} else if strings.HasSuffix(path, "/mj/insight-face/swap") {
		// midjourney plus
		relayMode = RelayModeSwapFace
	} else if strings.HasSuffix(path, "/submit/upload-discord-images") {
		// midjourney plus
		relayMode = RelayModeMidjourneyUpload
	} else if strings.HasSuffix(path, "/mj/submit/imagine") {
		relayMode = RelayModeMidjourneyImagine
	} else if strings.HasSuffix(path, "/mj/submit/video") {
		relayMode = RelayModeMidjourneyVideo
	} else if strings.HasSuffix(path, "/mj/submit/edits") {
		relayMode = RelayModeMidjourneyEdits
	} else if strings.HasSuffix(path, "/mj/submit/blend") {
		relayMode = RelayModeMidjourneyBlend
	} else if strings.HasSuffix(path, "/mj/submit/describe") {
		relayMode = RelayModeMidjourneyDescribe
	} else if strings.HasSuffix(path, "/mj/notify") {
		relayMode = RelayModeMidjourneyNotify
	} else if strings.HasSuffix(path, "/mj/submit/change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/mj/submit/simple-change") {
		relayMode = RelayModeMidjourneyChange
	} else if strings.HasSuffix(path, "/fetch") {
		relayMode = RelayModeMidjourneyTaskFetch
	} else if strings.HasSuffix(path, "/image-seed") {
		relayMode = RelayModeMidjourneyTaskImageSeed
	} else if strings.HasSuffix(path, "/list-by-condition") {
		relayMode = RelayModeMidjourneyTaskFetchByCondition
	}
	return relayMode
}

func Path2RelaySuno(method, path string) int {
	relayMode := RelayModeUnknown
	if method == http.MethodPost && strings.HasSuffix(path, "/fetch") {
		relayMode = RelayModeSunoFetch
	} else if method == http.MethodGet && strings.Contains(path, "/fetch/") {
		relayMode = RelayModeSunoFetchByID
	} else if strings.Contains(path, "/submit/") {
		relayMode = RelayModeSunoSubmit
	}
	return relayMode
}
