package vapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Dispatcher routes tool calls to implementations.
type Dispatcher struct {
	knowledge KnowledgeSearcher
	memory    MemoryStore
	github    GitHubSource
	jira      JiraSource
	slack     SlackSink
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(
	knowledge KnowledgeSearcher,
	memory MemoryStore,
	github GitHubSource,
	jira JiraSource,
	slack SlackSink,
) *Dispatcher {
	return &Dispatcher{
		knowledge: knowledge,
		memory:    memory,
		github:    github,
		jira:      jira,
		slack:     slack,
	}
}

// Dispatch routes a single tool call to its implementation.
func (d *Dispatcher) Dispatch(ctx context.Context, call ToolCall, userID string) ToolResult {
	result := ToolResult{ToolCallID: call.ID}

	switch call.Name {
	case "search_knowledge":
		query := mustString(call.Arguments, "query")
		topK := optInt(call.Arguments, "top_k", 5)
		hits, err := d.knowledge.SearchKnowledge(ctx, query, topK)
		if err != nil {
			result.Error = fmt.Sprintf("search failed: %v", err)
		} else {
			result.Result = formatSearchResults(hits)
		}

	case "recall_memory":
		query := mustString(call.Arguments, "query")
		topK := optInt(call.Arguments, "top_k", 3)
		items, err := d.memory.Recall(ctx, userID, query, topK)
		if err != nil {
			result.Error = fmt.Sprintf("recall failed: %v", err)
		} else {
			result.Result = strings.Join(items, "\n")
		}

	case "list_open_prs":
		author := optString(call.Arguments, "author", "")
		prs := d.github.ListOpenPRs(author)
		result.Result = formatPRList(prs)

	case "summarize_pr":
		prNum := mustInt(call.Arguments, "pr_number")
		pr, ok := d.github.GetPR(prNum)
		if !ok {
			result.Error = fmt.Sprintf("PR #%d not found", prNum)
		} else {
			result.Result = formatPRSummary(pr)
		}

	case "list_jira_tickets":
		assignee := optString(call.Arguments, "assignee", "")
		status := optString(call.Arguments, "status", "")
		tickets := d.jira.List(assignee, status)
		result.Result = formatTicketList(tickets)

	case "create_jira_ticket":
		summary := mustString(call.Arguments, "summary")
		description := mustString(call.Arguments, "description")
		priority := mustString(call.Arguments, "priority")
		ticket := d.jira.Create(summary, description, priority, userID)
		result.Result = formatTicketCreation(ticket)

	case "post_to_slack":
		channel := mustString(call.Arguments, "channel")
		message := mustString(call.Arguments, "message")
		msg, err := d.slack.Post(channel, userID, message)
		if err != nil {
			result.Error = fmt.Sprintf("slack post failed: %v", err)
		} else {
			result.Result = formatSlackPost(msg)
		}

	case "get_release_status":
		releaseName := optString(call.Arguments, "release_name", "default")
		status := d.aggregateReleaseStatus(releaseName)
		result.Result = status

	default:
		result.Error = fmt.Sprintf("unknown tool: %s", call.Name)
	}

	return result
}

// Helper functions for extracting arguments
func mustString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mustInt(args map[string]any, key string) int {
	if v, ok := args[key]; ok {
		switch tv := v.(type) {
		case float64:
			return int(tv)
		case int:
			return tv
		case string:
			if i, err := strconv.Atoi(tv); err == nil {
				return i
			}
		}
	}
	return 0
}

func optString(args map[string]any, key string, defaultVal string) string {
	if v := mustString(args, key); v != "" {
		return v
	}
	return defaultVal
}

func optInt(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch tv := v.(type) {
		case float64:
			return int(tv)
		case int:
			return tv
		case string:
			if i, err := strconv.Atoi(tv); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

// Formatting functions
func formatSearchResults(hits []SearchHit) string {
	if len(hits) == 0 {
		return "No results found."
	}
	var sb strings.Builder
	for i, hit := range hits {
		sb.WriteString(fmt.Sprintf("%d. %s (from %s, score: %.2f)\n%s\n\n",
			i+1, hit.Title, hit.Source, hit.Score, hit.Text))
	}
	return sb.String()
}

func formatPRList(prs []PRInfo) string {
	if len(prs) == 0 {
		return "No open PRs found."
	}
	var sb strings.Builder
	for _, pr := range prs {
		sb.WriteString(fmt.Sprintf("- #%d: %s (by %s, status: %s)\n",
			pr.Number, pr.Title, pr.Author, pr.Status))
	}
	return sb.String()
}

func formatPRSummary(pr *PRInfo) string {
	return fmt.Sprintf("PR #%d: %s\nAuthor: %s\nStatus: %s\nBody:\n%s",
		pr.Number, pr.Title, pr.Author, pr.Status, pr.Body)
}

func formatTicketList(tickets []TicketInfo) string {
	if len(tickets) == 0 {
		return "No tickets found."
	}
	var sb strings.Builder
	for _, t := range tickets {
		sb.WriteString(fmt.Sprintf("- %s: %s (status: %s, priority: %s, assignee: %s)\n",
			t.Key, t.Summary, t.Status, t.Priority, t.Assignee))
	}
	return sb.String()
}

func formatTicketCreation(ticket TicketInfo) string {
	return fmt.Sprintf("Created ticket %s: %s (priority: %s)", ticket.Key, ticket.Summary, ticket.Priority)
}

func formatSlackPost(msg MessageInfo) string {
	return fmt.Sprintf("Posted to #%s: %s", msg.Channel, msg.Text)
}

// aggregateReleaseStatus returns a summary of release blockers.
func (d *Dispatcher) aggregateReleaseStatus(releaseName string) string {
	openPRs := d.github.ListOpenPRs("")
	tickets := d.jira.List("", "")

	var blockers []string

	// Filter for high-priority/critical PRs
	for _, pr := range openPRs {
		if pr.Status == "open" {
			blockers = append(blockers, fmt.Sprintf("PR #%d: %s", pr.Number, pr.Title))
		}
	}

	// Filter for high/critical priority tickets with open status
	for _, t := range tickets {
		if (t.Status == "Open" || t.Status == "In Progress") &&
			(t.Priority == "High" || t.Priority == "Critical") {
			blockers = append(blockers, fmt.Sprintf("%s: %s", t.Key, t.Summary))
		}
	}

	if len(blockers) == 0 {
		return fmt.Sprintf("Release '%s' is clear. No open blockers.", releaseName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Release '%s' blockers:\n", releaseName))
	for i, b := range blockers {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, b))
	}
	return sb.String()
}
