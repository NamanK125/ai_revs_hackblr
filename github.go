// Package mocks provides in-memory mock implementations of external services.
package mocks

// PR represents a GitHub pull request.
type PR struct {
	Number int
	Title  string
	Author string
	Status string // "open", "merged", "closed"
	Body   string
	Files  []string
}

// GitHubMock is an in-memory mock of GitHub operations.
type GitHubMock struct {
	prs []PR
}

// NewGitHubMock creates a new GitHub mock with seeded data.
func NewGitHubMock() *GitHubMock {
	return &GitHubMock{
		prs: []PR{
			{
				Number: 234,
				Title:  "Fix login redirect race condition",
				Author: "alice",
				Status: "open",
				Body:   "Fixes the race condition in the auth service login redirect flow when multiple requests arrive simultaneously.",
				Files:  []string{"auth/login.go", "auth/session.go"},
			},
			{
				Number: 235,
				Title:  "Add Qdrant retry logic with exponential backoff",
				Author: "bob",
				Status: "open",
				Body:   "Implements retry logic with exponential backoff for transient failures in Qdrant calls.",
				Files:  []string{"internal/qdrant/client.go", "internal/qdrant/retry.go"},
			},
			{
				Number: 232,
				Title:  "Bump deepgram-sdk to v0.13.0",
				Author: "charlie",
				Status: "merged",
				Body:   "Updates deepgram-sdk to the latest version for better performance and bug fixes.",
				Files:  []string{"go.mod", "go.sum"},
			},
			{
				Number: 230,
				Title:  "Add observability: structured logging + metrics",
				Author: "alice",
				Status: "open",
				Body:   "Adds structured logging using slog and exports Prometheus metrics for key operations.",
				Files:  []string{"internal/log/log.go", "internal/metrics/metrics.go"},
			},
			{
				Number: 225,
				Title:  "Refactor memory.Manager interface",
				Author: "david",
				Status: "merged",
				Body:   "Simplifies the memory manager interface for better testability and clearer semantics.",
				Files:  []string{"internal/memory/memory.go", "internal/memory/memory_test.go"},
			},
			},
	}
}

// ListOpenPRs returns PRs matching the optional author filter.
// If author is empty, returns all PRs.
func (g *GitHubMock) ListOpenPRs(author string) []PR {
	var result []PR
	for _, pr := range g.prs {
		if pr.Status != "open" {
			continue
		}
		if author != "" && pr.Author != author {
			continue
		}
		result = append(result, pr)
	}
	return result
}

// GetPR retrieves a specific PR by number, returning (pr, found).
func (g *GitHubMock) GetPR(number int) (*PR, bool) {
	for i, pr := range g.prs {
		if pr.Number == number {
			return &g.prs[i], true
		}
	}
	return nil, false
}