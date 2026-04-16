package vapi

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Fake implementations for testing
type fakeKnowledge struct{}

func (f *fakeKnowledge) SearchKnowledge(ctx context.Context, query string, topK int) ([]SearchHit, error) {
	return []SearchHit{
		{
			Text:   "fake hit for " + query,
			Source: "stub.md",
			Title:  "Stub",
			Score:  0.99,
		},
	}, nil
}

type fakeMemory struct{}

func (f *fakeMemory) Recall(ctx context.Context, userID, query string, topK int) ([]string, error) {
	return []string{"(stub: no memory yet)"}, nil
}

func (f *fakeMemory) Write(ctx context.Context, userID, callID, summary string) error {
	return nil
}

type fakeGitHub struct{}

func (f *fakeGitHub) ListOpenPRs(author string) []PRInfo {
	return []PRInfo{
		{
			Number: 234,
			Title:  "Stub PR",
			Author: "alice",
			Status: "open",
			Body:   "stub body",
		},
	}
}

func (f *fakeGitHub) GetPR(number int) (*PRInfo, bool) {
	if number == 234 {
		return &PRInfo{
			Number: 234,
			Title:  "Stub PR",
			Author: "alice",
			Status: "open",
			Body:   "stub body",
		}, true
	}
	return nil, false
}

type fakeJira struct{}

func (f *fakeJira) List(assignee, status string) []TicketInfo {
	return []TicketInfo{
		{
			Key:      "ENG-1",
			Summary:  "Stub ticket",
			Status:   "Open",
			Assignee: "bob",
			Priority: "High",
		},
	}
}

func (f *fakeJira) Create(summary, description, priority, assignee string) TicketInfo {
	return TicketInfo{
		Key:      "ENG-999",
		Summary:  summary,
		Status:   "Open",
		Assignee: assignee,
		Priority: priority,
	}
}

type fakeSlack struct{}

func (f *fakeSlack) Post(channel, user, text string) (MessageInfo, error) {
	return MessageInfo{
		Channel:   channel,
		User:      user,
		Text:      text,
		Timestamp: 1681234567,
	}, nil
}

// TestToolWebhookSuccess tests a successful tool call.
func TestToolWebhookSuccess(t *testing.T) {
	// Create fake dependencies
	knowledge := &fakeKnowledge{}
	memory := &fakeMemory{}
	github := &fakeGitHub{}
	jira := &fakeJira{}
	slack := &fakeSlack{}

	dispatcher := NewDispatcher(knowledge, memory, github, jira, slack)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	handler := NewHandler(dispatcher, memory, "", logger)

	// Read testdata
	body, err := os.ReadFile("testdata/sample_tool_call.json")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	// Make request
	req := httptest.NewRequest("POST", "/vapi/tool", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ToolWebhook(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp ToolCallResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(resp.Results))
	}

	// Check first result (search_knowledge)
	if resp.Results[0].ToolCallID != "call-001" {
		t.Errorf("expected ToolCallID call-001, got %s", resp.Results[0].ToolCallID)
	}
	if resp.Results[0].Error != "" {
		t.Errorf("expected no error, got %s", resp.Results[0].Error)
	}

	// Check second result (create_jira_ticket)
	if resp.Results[1].ToolCallID != "call-002" {
		t.Errorf("expected ToolCallID call-002, got %s", resp.Results[1].ToolCallID)
	}
	if resp.Results[1].Error != "" {
		t.Errorf("expected no error, got %s", resp.Results[1].Error)
	}
}

// TestToolWebhookSecretVerification tests that wrong secret is rejected.
func TestToolWebhookSecretVerification(t *testing.T) {
	knowledge := &fakeKnowledge{}
	memory := &fakeMemory{}
	github := &fakeGitHub{}
	jira := &fakeJira{}
	slack := &fakeSlack{}

	dispatcher := NewDispatcher(knowledge, memory, github, jira, slack)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	handler := NewHandler(dispatcher, memory, "correct-secret", logger)

	body, err := os.ReadFile("testdata/sample_tool_call.json")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}

	req := httptest.NewRequest("POST", "/vapi/tool", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-vapi-secret", "wrong-secret")
	w := httptest.NewRecorder()

	handler.ToolWebhook(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
