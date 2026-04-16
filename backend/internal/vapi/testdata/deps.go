package vapi

import "context"

// KnowledgeSearcher searches semantic knowledge.
type KnowledgeSearcher interface {
	SearchKnowledge(ctx context.Context, query string, topK int) ([]SearchHit, error)
}

// SearchHit is a single search result.
type SearchHit struct {
	Text   string
	Source string
	Title  string
	Score  float32
}

// MemoryStore handles episodic memory storage and retrieval.
type MemoryStore interface {
	Recall(ctx context.Context, userID, query string, topK int) ([]string, error)
	Write(ctx context.Context, userID, callID, summary string) error
}

// GitHubSource provides GitHub data.
type GitHubSource interface {
	ListOpenPRs(author string) []PRInfo
	GetPR(number int) (*PRInfo, bool)
}

// PRInfo contains GitHub PR metadata.
type PRInfo struct {
	Number int
	Title  string
	Author string
	Status string
	Body   string
}

// JiraSource provides Jira operations.
type JiraSource interface {
	List(assignee, status string) []TicketInfo
	Create(summary, description, priority, assignee string) TicketInfo
}

// TicketInfo contains Jira ticket metadata.
type TicketInfo struct {
	Key      string
	Summary  string
	Status   string
	Assignee string
	Priority string
}

// SlackSink posts to Slack.
type SlackSink interface {
	Post(channel, user, text string) (MessageInfo, error)
}

// MessageInfo contains Slack message metadata.
type MessageInfo struct {
	Channel   string
	User      string
	Text      string
	Timestamp int64
}
