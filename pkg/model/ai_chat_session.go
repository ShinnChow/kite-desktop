package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type AIChatSession struct {
	Model
	SessionID     string    `json:"sessionId" gorm:"type:varchar(64);uniqueIndex;not null"`
	Title         string    `json:"title" gorm:"type:varchar(255);not null"`
	ClusterName   string    `json:"clusterName" gorm:"type:varchar(100);not null;index:idx_ai_chat_sessions_cluster_updated,priority:1"`
	Page          string    `json:"page" gorm:"type:varchar(100)"`
	Namespace     string    `json:"namespace" gorm:"type:varchar(100)"`
	ResourceName  string    `json:"resourceName" gorm:"type:varchar(255)"`
	ResourceKind  string    `json:"resourceKind" gorm:"type:varchar(100)"`
	MessageCount  int       `json:"messageCount" gorm:"not null;default:0"`
	LastMessageAt time.Time `json:"lastMessageAt" gorm:"index:idx_ai_chat_sessions_cluster_updated,priority:2,sort:desc"`
}

type AIChatMessage struct {
	Model
	SessionID     string    `json:"sessionId" gorm:"type:varchar(64);not null;index:idx_ai_chat_messages_session_seq,priority:1;uniqueIndex:idx_ai_chat_messages_session_message,priority:1"`
	MessageID     string    `json:"messageId" gorm:"type:varchar(64);not null;uniqueIndex:idx_ai_chat_messages_session_message,priority:2"`
	Seq           int       `json:"seq" gorm:"not null;uniqueIndex:idx_ai_chat_messages_session_seq,priority:2"`
	Role          string    `json:"role" gorm:"type:varchar(20);not null"`
	Content       string    `json:"content" gorm:"type:text"`
	Thinking      string    `json:"thinking" gorm:"type:text"`
	ToolCallID    string    `json:"toolCallId" gorm:"type:varchar(255)"`
	ToolName      string    `json:"toolName" gorm:"type:varchar(255)"`
	ToolArgs      JSONField `json:"toolArgs" gorm:"type:text"`
	ToolResult    string    `json:"toolResult" gorm:"type:text"`
	ActionStatus  string    `json:"actionStatus" gorm:"type:varchar(20)"`
	InputRequest  JSONField `json:"inputRequest" gorm:"type:text"`
	PendingAction JSONField `json:"pendingAction" gorm:"type:text"`
}

type AIChatSessionSnapshot struct {
	SessionID    string
	Title        string
	ClusterName  string
	Page         string
	Namespace    string
	ResourceName string
	ResourceKind string
	Messages     []AIChatMessage
}

func ListAIChatSessions(clusterName string, page, pageSize int) ([]AIChatSession, int64, error) {
	query := DB.Model(&AIChatSession{})
	if clusterName != "" {
		query = query.Where("cluster_name = ?", clusterName)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sessions := make([]AIChatSession, 0, pageSize)
	if err := query.
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func GetAIChatSession(clusterName, sessionID string) (*AIChatSession, error) {
	var session AIChatSession
	query := DB.Where("session_id = ?", sessionID)
	if clusterName != "" {
		query = query.Where("cluster_name = ?", clusterName)
	}
	if err := query.First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func ListAIChatMessages(sessionID string) ([]AIChatMessage, error) {
	messages := make([]AIChatMessage, 0)
	if err := DB.
		Where("session_id = ?", sessionID).
		Order("seq ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func UpsertAIChatSession(snapshot AIChatSessionSnapshot) (*AIChatSession, error) {
	if snapshot.SessionID == "" {
		return nil, errors.New("session id is required")
	}

	historyLimit := DefaultAIChatHistorySessionLimit
	if setting, err := GetGeneralSetting(); err == nil {
		historyLimit = NormalizeAIChatHistorySessionLimit(setting.AIChatHistorySessionLimit)
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		var session AIChatSession
		err := tx.Where("session_id = ?", snapshot.SessionID).First(&session).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		updates := map[string]interface{}{
			"title":           snapshot.Title,
			"cluster_name":    snapshot.ClusterName,
			"page":            snapshot.Page,
			"namespace":       snapshot.Namespace,
			"resource_name":   snapshot.ResourceName,
			"resource_kind":   snapshot.ResourceKind,
			"message_count":   len(snapshot.Messages),
			"last_message_at": now,
			"updated_at":      now,
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			session = AIChatSession{
				SessionID:     snapshot.SessionID,
				Title:         snapshot.Title,
				ClusterName:   snapshot.ClusterName,
				Page:          snapshot.Page,
				Namespace:     snapshot.Namespace,
				ResourceName:  snapshot.ResourceName,
				ResourceKind:  snapshot.ResourceKind,
				MessageCount:  len(snapshot.Messages),
				LastMessageAt: now,
			}
			if createErr := tx.Create(&session).Error; createErr != nil {
				return createErr
			}
		} else {
			if updateErr := tx.Model(&session).Updates(updates).Error; updateErr != nil {
				return updateErr
			}
			if refreshErr := tx.Where("session_id = ?", snapshot.SessionID).First(&session).Error; refreshErr != nil {
				return refreshErr
			}
		}

		existingMessages := make([]AIChatMessage, 0, len(snapshot.Messages))
		if listErr := tx.Where("session_id = ?", snapshot.SessionID).Find(&existingMessages).Error; listErr != nil {
			return listErr
		}

		existingByMessageID := make(map[string]AIChatMessage, len(existingMessages))
		for _, message := range existingMessages {
			existingByMessageID[message.MessageID] = message
		}

		messagesToDelete := make([]AIChatMessage, 0)
		for _, message := range existingMessages {
			if containsMessageID(snapshot.Messages, message.MessageID) {
				continue
			}
			messagesToDelete = append(messagesToDelete, message)
		}
		for _, message := range messagesToDelete {
			if deleteErr := tx.Delete(&message).Error; deleteErr != nil {
				return deleteErr
			}
			delete(existingByMessageID, message.MessageID)
		}

		tempSeqOffset := len(snapshot.Messages) + len(existingMessages) + 1
		for idx, message := range snapshot.Messages {
			existing, ok := existingByMessageID[message.MessageID]
			if !ok {
				continue
			}
			targetSeq := idx + 1
			if existing.Seq == targetSeq {
				continue
			}
			if updateErr := tx.Model(&existing).Update("seq", targetSeq+tempSeqOffset).Error; updateErr != nil {
				return updateErr
			}
		}

		for _, message := range snapshot.Messages {
			message.SessionID = snapshot.SessionID

			if existing, ok := existingByMessageID[message.MessageID]; ok {
				updates := map[string]interface{}{
					"seq":            message.Seq,
					"role":           message.Role,
					"content":        message.Content,
					"thinking":       message.Thinking,
					"tool_call_id":   message.ToolCallID,
					"tool_name":      message.ToolName,
					"tool_args":      message.ToolArgs,
					"tool_result":    message.ToolResult,
					"action_status":  message.ActionStatus,
					"input_request":  message.InputRequest,
					"pending_action": message.PendingAction,
				}
				if updateErr := tx.Model(&existing).Updates(updates).Error; updateErr != nil {
					return updateErr
				}
				continue
			}

			if createErr := tx.Create(&message).Error; createErr != nil {
				return createErr
			}
		}

		if cleanupErr := cleanupAIChatSessions(tx, snapshot.ClusterName, historyLimit); cleanupErr != nil {
			return cleanupErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return GetAIChatSession(snapshot.ClusterName, snapshot.SessionID)
}

func DeleteAIChatSession(clusterName, sessionID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var session AIChatSession
		query := tx.Where("session_id = ?", sessionID)
		if clusterName != "" {
			query = query.Where("cluster_name = ?", clusterName)
		}
		if err := query.First(&session).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ?", sessionID).Delete(&AIChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&session).Error
	})
}

func cleanupAIChatSessions(tx *gorm.DB, clusterName string, limit int) error {
	normalizedLimit := NormalizeAIChatHistorySessionLimit(limit)
	if clusterName == "" {
		return nil
	}

	var expiredSessionIDs []string
	if err := tx.Model(&AIChatSession{}).
		Where("cluster_name = ?", clusterName).
		Order("updated_at DESC").
		Offset(normalizedLimit).
		Pluck("session_id", &expiredSessionIDs).Error; err != nil {
		return err
	}
	if len(expiredSessionIDs) == 0 {
		return nil
	}

	if err := tx.Where("session_id IN ?", expiredSessionIDs).Delete(&AIChatMessage{}).Error; err != nil {
		return err
	}
	return tx.Where("session_id IN ?", expiredSessionIDs).Delete(&AIChatSession{}).Error
}

func containsMessageID(messages []AIChatMessage, messageID string) bool {
	for _, message := range messages {
		if message.MessageID == messageID {
			return true
		}
	}
	return false
}
