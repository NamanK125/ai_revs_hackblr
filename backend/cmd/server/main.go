package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihup/engineerops/internal/vapi"
)

// Need to forward-declare context for fakes since they use it
var _ context.Context

func main() {
	// Setup logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Read environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	vapiSecret := os.Getenv("VAPI_SHARED_SECRET")

	// Create fake implementations
	knowledge := &fakeKnowledge{}
	memory := &fakeMemory{}
	github := &fakeGitHub{}
	jira := &fakeJira{}
	slack := &fakeSlack{}

	// Wire up Dispatcher and Handler
	dispatcher := vapi.NewDispatcher(knowledge, memory, github, jira, slack)
	handler := vapi.NewHandler(dispatcher, memory, vapiSecret, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /vapi/tool", handler.ToolWebhook)
	mux.HandleFunc("POST /vapi/end-of-call", handler.EndOfCallWebhook)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

// Fake implementations of vapi interfaces

type fakeKnowledge struct{}

func (f *fakeKnowledge) SearchKnowledge(ctx context.Context, query string, topK int) ([]vapi.SearchHit, error) {
	return []vapi.SearchHit{
		{
			Text:   "This is stub knowledge about " + query,
			Source: "stub.md",
			Title:  "Stub Knowledge",
			Score:  0.99,
		},
	}, nil
}

type fakeMemory struct{}

func (f *fakeMemory) Recall(ctx context.Context, userID, query string, topK int) ([]string, error) {
	return []string{"(stub: no memory yet)"}, nil
}

func (f *fakeMemory) Write(ctx context.Context, userID, callID, summary string) error {
	slog.Info("memory write", "user", userID, "call_id", callID, "summary", summary)
	return nil
}

type fakeGitHub struct{}

func (f *fakeGitHub) ListOpenPRs(author string) []vapi.PRInfo {
	return []vapi.PRInfo{
		{
			Number: 234,
			Title:  "Stub PR - implement feature X",
			Author: "alice",
			Status: "open",
			Body:   "This is a stub PR body with some details about feature X",
		},
	}
}

func (f *fakeGitHub) GetPR(number int) (*vapi.PRInfo, bool) {
	if number == 234 {
		return &vapi.PRInfo{
			Number: 234,
			Title:  "Stub PR - implement feature X",
			Author: "alice",
			Status: "open",
			Body:   "This is a stub PR body with some details about feature X",
		}, true
	}
	return nil, false
}

type fakeJira struct{}

func (f *fakeJira) List(assignee, status string) []vapi.TicketInfo {
	return []vapi.TicketInfo{
		{
			Key:      "ENG-1",
			Summary:  "Stub ticket - fix auth bug",
			Status:   "Open",
			Assignee: "bob",
			Priority: "High",
		},
	}
}

func (f *fakeJira) Create(summary, description, priority, assignee string) vapi.TicketInfo {
	return vapi.TicketInfo{
		Key:      "ENG-999",
		Summary:  summary,
		Status:   "Open",
		Assignee: assignee,
		Priority: priority,
	}
}

type fakeSlack struct{}

func (f *fakeSlack) Post(channel, user, text string) (vapi.MessageInfo, error) {
	return vapi.MessageInfo{
		Channel:   channel,
		User:      user,
		Text:      text,
		Timestamp: time.Now().Unix(),
	}, nil
}
