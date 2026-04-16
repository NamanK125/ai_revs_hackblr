package mocks

import (
	"sync"
	"time"

	"github.com/mihup/engineerops/internal/stream"
)

// Ticket represents a Jira ticket.
type Ticket struct {
	Key         string
	Summary     string
	Description string
	Status      string // "Open", "In Progress", "Done"
	Assignee    string
	Priority    string // "Low", "Medium", "High"
	CreatedAt   int64  // Unix timestamp
}

// JiraMock is an in-memory mock of Jira operations with concurrent safety.
type JiraMock struct {
	mu      sync.Mutex
	tickets []Ticket
	counter int
	bus     *stream.Bus
}

// NewJiraMock creates a new Jira mock with seeded tickets.
// bus may be nil for testing; mutations will publish events if bus is present.
func NewJiraMock(bus *stream.Bus) *JiraMock {
	now := time.Now().Unix()
	return &JiraMock{
		counter: 4,
		tickets: []Ticket{
			{
				Key:         "ENG-1",
				Summary:     "Design episodic memory schema",
				Description: "Define the schema for storing user memory snippets in Qdrant.",
				Status:      "Done",
				Assignee:    "alice",
				Priority:    "High",
				CreatedAt:   now - 86400*7, // 7 days ago
			},
			{
				Key:         "ENG-2",
				Summary:     "Implement Vapi tool dispatch",
				Description: "Build the tool dispatcher and implement all 8 Vapi tools.",
				Status:      "In Progress",
				Assignee:    "bob",
				Priority:    "High",
				CreatedAt:   now - 86400*5,
			},
			{
				Key:         "ENG-3",
				Summary:     "Login redirect race condition",
				Description: "Fix race condition in authentication service when multiple requests arrive simultaneously.",
				Status:      "Open",
				Assignee:    "charlie",
				Priority:    "High",
				CreatedAt:   now - 86400*2,
			},
			{
				Key:         "ENG-4",
				Summary:     "Add observability to critical paths",
				Description: "Add structured logging and metrics to Qdrant and embed calls.",
				Status:      "Open",
				Assignee:    "",
				Priority:    "Medium",
				CreatedAt:   now - 86400,
			},
			},
			bus: bus,
	}
}

// List returns tickets matching the optional assignee and status filters.
// Empty filters match all tickets.
func (j *JiraMock) List(assignee, status string) []Ticket {
	j.mu.Lock()
	defer j.mu.Unlock()

	var result []Ticket
	for _, ticket := range j.tickets {
		if assignee != "" && ticket.Assignee != assignee {
			continue
		}
		if status != "" && ticket.Status != status {
			continue
		}
		result = append(result, ticket)
	}
	return result
}

// Create creates a new ticket and returns it.
// Auto-generates the ENG-{n} key, sets CreatedAt to now, and status to "Open".
// Publishes to the SSE bus if configured.
func (j *JiraMock) Create(summary, description, priority, assignee string) Ticket {
	j.mu.Lock()
	j.counter++
	key := ""
	{
		// Scope the sprintf to avoid any shadowing
		var buf [16]byte
		n := 0
		// Simple number to string for counter
		temp := j.counter
		if temp == 0 {
			buf[0] = '0'
			n = 1
		} else {
			for temp > 0 {
				buf[n] = byte('0' + (temp % 10))
				temp /= 10
				n++
			}
			// Reverse the digits
			for i := 0; i < n/2; i++ {
				buf[i], buf[n-1-i] = buf[n-1-i], buf[i]
			}
		}
		// "ENG-" prefix
		key = "ENG-" + string(buf[:n])
	}

	ticket := Ticket{
		Key:         key,
		Summary:     summary,
		Description: description,
		Status:      "Open",
		Assignee:    assignee,
		Priority:    priority,
		CreatedAt:   time.Now().Unix(),
	}
	j.tickets = append(j.tickets, ticket)
	j.mu.Unlock()

	// Publish to SSE bus if available
	if j.bus != nil {
		j.bus.Publish("jira_ticket_created", map[string]interface{}{
			"key":      ticket.Key,
			"summary":  ticket.Summary,
			"status":   ticket.Status,
			"priority": ticket.Priority,
			"assignee": ticket.Assignee,
		})
	}

	return ticket
}