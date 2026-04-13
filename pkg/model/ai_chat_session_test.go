package model

import (
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"
)

func TestAIChatSessionCRUD(t *testing.T) {
	if err := DB.Where("session_id LIKE ?", "chat-session-%").Delete(&AIChatMessage{}).Error; err != nil {
		t.Fatalf("cleanup messages error = %v", err)
	}
	if err := DB.Where("session_id LIKE ?", "chat-session-%").Delete(&AIChatSession{}).Error; err != nil {
		t.Fatalf("cleanup sessions error = %v", err)
	}

	sessionID := "chat-session-1"
	snapshot := AIChatSessionSnapshot{
		SessionID:    sessionID,
		Title:        "Investigate pod restart",
		ClusterName:  "cluster-a",
		Page:         "pod-detail",
		Namespace:    "default",
		ResourceName: "nginx",
		ResourceKind: "pod",
		Messages: []AIChatMessage{
			{SessionID: sessionID, MessageID: "m1", Seq: 1, Role: "user", Content: "why restart"},
			{SessionID: sessionID, MessageID: "m2", Seq: 2, Role: "assistant", Content: "checking"},
		},
	}

	saved, err := UpsertAIChatSession(snapshot)
	if err != nil {
		t.Fatalf("UpsertAIChatSession() error = %v", err)
	}
	if saved.SessionID != sessionID {
		t.Fatalf("SessionID = %q, want %q", saved.SessionID, sessionID)
	}
	if saved.MessageCount != 2 {
		t.Fatalf("MessageCount = %d, want 2", saved.MessageCount)
	}

	sessions, total, err := ListAIChatSessions("cluster-a", 1, 20)
	if err != nil {
		t.Fatalf("ListAIChatSessions() error = %v", err)
	}
	if total < 1 {
		t.Fatalf("total = %d, want >= 1", total)
	}
	if len(sessions) == 0 {
		t.Fatal("sessions is empty")
	}
	if sessions[0].SessionID != sessionID {
		t.Fatalf("sessions[0].SessionID = %q, want %q", sessions[0].SessionID, sessionID)
	}

	messages, err := ListAIChatMessages(sessionID)
	if err != nil {
		t.Fatalf("ListAIChatMessages() error = %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(messages))
	}
	if messages[0].MessageID != "m1" || messages[1].MessageID != "m2" {
		t.Fatalf("unexpected message ids: %#v", messages)
	}
	firstMessageRowID := messages[0].ID

	updated, err := UpsertAIChatSession(AIChatSessionSnapshot{
		SessionID:    sessionID,
		Title:        "Investigate pod restart updated",
		ClusterName:  "cluster-a",
		Page:         "pod-detail",
		Namespace:    "default",
		ResourceName: "nginx",
		ResourceKind: "pod",
		Messages: []AIChatMessage{
			{SessionID: sessionID, MessageID: "m1", Seq: 1, Role: "user", Content: "why restart"},
			{SessionID: sessionID, MessageID: "m3", Seq: 2, Role: "assistant", Content: "new answer"},
			{SessionID: sessionID, MessageID: "m4", Seq: 3, Role: "tool", Content: "tool"},
		},
	})
	if err != nil {
		t.Fatalf("second UpsertAIChatSession() error = %v", err)
	}
	if updated.Title != "Investigate pod restart updated" {
		t.Fatalf("Title = %q, want updated title", updated.Title)
	}
	if updated.MessageCount != 3 {
		t.Fatalf("MessageCount = %d, want 3", updated.MessageCount)
	}

	messages, err = ListAIChatMessages(sessionID)
	if err != nil {
		t.Fatalf("ListAIChatMessages() after update error = %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("len(messages) after update = %d, want 3", len(messages))
	}
	if messages[0].ID != firstMessageRowID {
		t.Fatalf("messages[0].ID = %d, want unchanged row id %d", messages[0].ID, firstMessageRowID)
	}
	if messages[1].MessageID != "m3" {
		t.Fatalf("messages[1].MessageID = %q, want %q", messages[1].MessageID, "m3")
	}

	if err := DeleteAIChatSession("cluster-a", sessionID); err != nil {
		t.Fatalf("DeleteAIChatSession() error = %v", err)
	}

	if _, err := GetAIChatSession("cluster-a", sessionID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("GetAIChatSession() error = %v, want record not found", err)
	}
	messages, err = ListAIChatMessages(sessionID)
	if err != nil {
		t.Fatalf("ListAIChatMessages() after delete error = %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("len(messages) after delete = %d, want 0", len(messages))
	}
}

func TestAIChatSessionRetentionLimit(t *testing.T) {
	if _, err := GetGeneralSetting(); err != nil {
		t.Fatalf("GetGeneralSetting() error = %v", err)
	}
	if err := DB.Model(&GeneralSetting{}).Where("id = ?", 1).Update("ai_chat_history_session_limit", 2).Error; err != nil {
		t.Fatalf("set history limit error = %v", err)
	}
	t.Cleanup(func() {
		if err := DB.Model(&GeneralSetting{}).Where("id = ?", 1).Update("ai_chat_history_session_limit", DefaultAIChatHistorySessionLimit).Error; err != nil {
			t.Fatalf("restore history limit error = %v", err)
		}
	})

	if err := DB.Where("session_id LIKE ?", "retention-session-%").Delete(&AIChatMessage{}).Error; err != nil {
		t.Fatalf("cleanup messages error = %v", err)
	}
	if err := DB.Where("session_id LIKE ?", "retention-session-%").Delete(&AIChatSession{}).Error; err != nil {
		t.Fatalf("cleanup sessions error = %v", err)
	}

	for i := 1; i <= 3; i++ {
		sessionID := fmt.Sprintf("retention-session-%d", i)
		_, err := UpsertAIChatSession(AIChatSessionSnapshot{
			SessionID:   sessionID,
			Title:       sessionID,
			ClusterName: "cluster-retention",
			Page:        "overview",
			Messages: []AIChatMessage{
				{SessionID: sessionID, MessageID: "m1", Seq: 1, Role: "user", Content: sessionID},
			},
		})
		if err != nil {
			t.Fatalf("UpsertAIChatSession(%s) error = %v", sessionID, err)
		}
	}

	sessions, _, err := ListAIChatSessions("cluster-retention", 1, 10)
	if err != nil {
		t.Fatalf("ListAIChatSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(sessions))
	}
	for _, session := range sessions {
		if session.SessionID == "retention-session-1" {
			t.Fatalf("oldest session was not cleaned up: %#v", sessions)
		}
	}

	messages, err := ListAIChatMessages("retention-session-1")
	if err != nil {
		t.Fatalf("ListAIChatMessages() error = %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("retained messages for deleted session: %#v", messages)
	}
}
