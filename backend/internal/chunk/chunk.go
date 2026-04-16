// Package chunk provides utilities for splitting markdown documents into chunks.
package chunk

// Chunk represents a piece of text extracted from a source document.
type Chunk struct {
	Text        string // The actual chunk content
	SourcePath  string // Path to the source file
	Title       string // Title of the document
	HeadingPath string // Breadcrumb path of headings (e.g., "Deployment > Staging > Health Checks")
	Index       int    // Sequential index of this chunk in the source
}

// MarkdownChunks splits markdown content into chunks.
// Will be replaced in Task 2 with H2/H3-aware splitter + token-window overlap.
// Currently returns a single chunk with the whole content as a stub.

// Parameters:
//   - sourcePath: path to the source file (for citation)
//   - title: document title
//   - content: full markdown content
//   - targetTokens: target chunk size in tokens (approximation)
//   - overlapTokens: overlap between consecutive chunks in tokens

// Returns a list of chunks.
func MarkdownChunks(sourcePath, title, content string, targetTokens int, overlapTokens int) []Chunk {
	// TODO: Implement H2/H3-aware splitter with token-window overlap in Task 2
	// For now, return single chunk with whole content
	_ = targetTokens
	_ = overlapTokens
	return []Chunk{
		{
			Text:        content,
			SourcePath:  sourcePath,
			Title:       title,
			HeadingPath: "",
			Index:       0,
		},
	}
}

// EstimateTokens estimates the number of tokens in a string.
// Uses the heuristic that 1 token ≈ 4 characters.
func EstimateTokens(s string) int {
	return len(s) / 4
}