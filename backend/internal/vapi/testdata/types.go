package vapi

// ToolCallRequest represents a Vapi webhook for tool calls.
type ToolCallRequest struct {
	Message struct {
		Type         string     `json:"type"` // "tool-calls"
		Timestamp    int64      `json:"timestamp"`
		ToolCallList []ToolCall `json:"toolCallList"`
		Call         *Call      `json:"call,omitempty"`
	} `json:"message"`
}

// ToolCall is a single tool invocation from Vapi.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Call contains metadata about the Vapi conversation.
type Call struct {
	ID                 string `json:"id"`
	AssistantOverrides *struct {
		VariableValues map[string]any `json:"variableValues"`
	} `json:"assistantOverrides,omitempty"`
}

// ToolResult is the result of one tool call to return to Vapi.
type ToolResult struct {
	ToolCallID string `json:"toolCallId"`
	Result     any    `json:"result"`
	Error      string `json:"error,omitempty"`
}

// ToolCallResponse is sent back to Vapi with all tool results.
type ToolCallResponse struct {
	Results []ToolResult `json:"results"`
}

// EndOfCallRequest represents the end-of-call-report webhook.
type EndOfCallRequest struct {
	Message struct {
		Type        string `json:"type"` // "end-of-call-report"
		Call        Call   `json:"call"`
		Summary     string `json:"summary"`
		Transcript  string `json:"transcript"`
		EndedReason string `json:"endedReason"`
	} `json:"message"`
}
