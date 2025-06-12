package event

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// FocalboardEvent 代表从 Webhook 接收到的原始事件结构
// 注意：Focalboard 的 Webhook 结构可能更复杂，这里是一个简化的通用模型
// 您可以根据实际的 payload 调整此结构
type FocalboardEvent struct {
	Type      string          `json:"type"`      // 例如 "block-changed"
	UserID    string          `json:"userId"`    // 操作用户ID
	BlockID   string          `json:"blockId"`   // 卡片/块ID
	BoardID   string          `json:"boardId"`   // 看板ID
	Timestamp int64           `json:"timestamp"` // 事件时间戳
	Payload   json.RawMessage `json:"payload"`   // 完整的原始 payload
}

// StoredEvent 存储在内存中的事件信息
type StoredEvent struct {
	FirstSeen time.Time       `json:"firstSeen"`
	LastSeen  time.Time       `json:"lastSeen"`
	Count     int64           `json:"count"`
	Details   FocalboardEvent `json:"details"`
}

// EventStore 是一个线程安全的内存事件存储
type EventStore struct {
	mu     sync.Mutex
	events map[string]*StoredEvent
}

// NewEventStore 创建一个新的 EventStore 实例
func NewEventStore() *EventStore {
	return &EventStore{
		events: make(map[string]*StoredEvent),
	}
}

// generateKey 根据事件的核心内容生成一个唯一的键
// 我们使用 Type, UserID, BlockID, BoardID 的组合
func (e *FocalboardEvent) generateKey() string {
	// 为了简化，我们使用格式化字符串。对于复杂 payload，可以考虑哈希。
	key := fmt.Sprintf("type:%s|user:%s|block:%s|board:%s", e.Type, e.UserID, e.BlockID, e.BoardID)
	// 使用 SHA256 确保键的长度一致且安全
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// Process 处理一个新接收到的事件
func (s *EventStore) Process(event FocalboardEvent) *StoredEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := event.generateKey()
	now := time.Now()

	if existingEvent, ok := s.events[key]; ok {
		// 事件已存在，计数器+1
		existingEvent.Count++
		existingEvent.LastSeen = now
		return existingEvent
	}

	// 新事件，创建记录
	newEvent := &StoredEvent{
		FirstSeen: now,
		LastSeen:  now,
		Count:     1,
		Details:   event,
	}
	s.events[key] = newEvent
	return newEvent
}

// LogEvent 使用 slog 记录事件处理结果
func LogEvent(processedEvent *StoredEvent) {
	slog.Info("Focalboard event processed",
		"eventType", processedEvent.Details.Type,
		"count", processedEvent.Count,
		"eventDetails", processedEvent.Details, // slog 会自动处理结构体
	)
}
