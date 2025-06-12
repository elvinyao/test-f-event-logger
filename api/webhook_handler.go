package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/elvinyao/test-f-event-logger/event"
)

// WebhookHandler 封装了处理 Webhook 所需的依赖
type WebhookHandler struct {
	store     *event.EventStore
	authToken string
}

// NewWebhookHandler 创建一个新的 WebhookHandler
func NewWebhookHandler(store *event.EventStore, token string) *WebhookHandler {
	return &WebhookHandler{
		store:     store,
		authToken: token,
	}
}

// HandleWebhook 是处理 /webhook 路由的 http.HandlerFunc
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 验证认证令牌
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != h.authToken {
		slog.Warn("Unauthorized webhook attempt", "remoteAddr", r.RemoteAddr, "token_provided", token)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 3. 读取和解析请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var fbEvent event.FocalboardEvent
	if err := json.Unmarshal(body, &fbEvent); err != nil {
		slog.Error("Failed to unmarshal JSON payload", "error", err, "body", string(body))
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}
	// 将原始 payload 也存入，方便日志记录
	fbEvent.Payload = body

	// 4. 处理事件
	processedEvent := h.store.Process(fbEvent)

	// 5. 记录日志
	event.LogEvent(processedEvent)

	// 6. 响应成功
	w.WriteHeader(http.StatusNoContent)
}
