package vapi

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Handler handles Vapi webhooks.
type Handler struct {
	dispatcher   *Dispatcher
	memory       MemoryStore
	sharedSecret string
	logger       *slog.Logger
}

// NewHandler creates a new Vapi webhook handler.
func NewHandler(d *Dispatcher, mem MemoryStore, sharedSecret string, logger *slog.Logger) *Handler {
	return &Handler{
		dispatcher:   d,
		memory:       mem,
		sharedSecret: sharedSecret,
		logger:       logger,
	}
}

// ToolWebhook handles POST /vapi/tool.
func (h *Handler) ToolWebhook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Verify secret if configured
	if h.sharedSecret != "" {
		if !h.verifySecret(r) {
			h.logger.Warn("tool webhook secret verification failed",
				"remote_addr", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
	} else {
		h.logger.Warn("no shared secret configured; skipping verification")
	}

	// Parse request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req ToolCallRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to unmarshal request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extract userID
	userID := "anonymous"
	if req.Message.Call != nil && req.Message.Call.AssistantOverrides != nil {
		if uid, ok := req.Message.Call.AssistantOverrides.VariableValues["userId"]; ok {
			if str, ok := uid.(string); ok {
				userID = str
			}
		}
	}

	// Dispatch all tool calls concurrently
	var wg sync.WaitGroup
	results := make([]ToolResult, len(req.Message.ToolCallList))
	resultsMu := sync.Mutex{}

	for i, toolCall := range req.Message.ToolCallList {
		wg.Add(1)
		go func(idx int, call ToolCall) {
			defer wg.Done()
			result := h.dispatcher.Dispatch(ctx, call, userID)
			resultsMu.Lock()
			results[idx] = result
			resultsMu.Unlock()
		}(i, toolCall)
	}

	wg.Wait()

	// Build response
	resp := ToolCallResponse{Results: results}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	// Log
	latencyMs := time.Since(start).Milliseconds()
	h.logger.Info("tool webhook completed",
		"user", userID,
		"tool_count", len(req.Message.ToolCallList),
		"latency_ms", latencyMs)
}

// EndOfCallWebhook handles POST /vapi/end-of-call.
func (h *Handler) EndOfCallWebhook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Verify secret if configured
	if h.sharedSecret != "" {
		if !h.verifySecret(r) {
			h.logger.Warn("end-of-call webhook secret verification failed",
				"remote_addr", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
	}

	// Parse request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req EndOfCallRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("failed to unmarshal end-of-call request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extract userID and callID
	userID := "anonymous"
	if req.Message.Call.AssistantOverrides != nil {
		if uid, ok := req.Message.Call.AssistantOverrides.VariableValues["userId"]; ok {
			if str, ok := uid.(string); ok {
				userID = str
			}
		}
	}

	callID := req.Message.Call.ID
	summary := req.Message.Summary

	// Write to memory store
	if err := h.memory.Write(ctx, userID, callID, summary); err != nil {
		h.logger.Error("failed to write to memory",
			"user", userID,
			"call_id", callID,
			"error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	// Log
	latencyMs := time.Since(start).Milliseconds()
	h.logger.Info("end-of-call webhook completed",
		"user", userID,
		"call_id", callID,
		"latency_ms", latencyMs)
}

// verifySecret performs constant-time comparison of x-vapi-secret header.
func (h *Handler) verifySecret(r *http.Request) bool {
	headerValue := r.Header.Get("x-vapi-secret")
	return subtle.ConstantTimeCompare([]byte(headerValue), []byte(h.sharedSecret)) == 1
}
