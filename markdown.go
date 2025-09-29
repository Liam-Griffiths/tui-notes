package main

import (
	"regexp"
	"strings"
)

// =============================================================================
// MARKDOWN RENDERING
// =============================================================================

// renderMarkdown converts markdown content to display format
func (app *App) renderMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		result = append(result, app.renderMarkdownLine(line))
	}

	return strings.Join(result, "\n")
}

// renderMarkdownLine processes a single line of markdown
func (app *App) renderMarkdownLine(line string) string {
	// Handle headers
	if strings.HasPrefix(line, "# ") {
		return "🔸 " + strings.TrimPrefix(line, "# ")
	}
	if strings.HasPrefix(line, "## ") {
		return "  • " + strings.TrimPrefix(line, "## ")
	}
	if strings.HasPrefix(line, "### ") {
		return "    ◦ " + strings.TrimPrefix(line, "### ")
	}
	if strings.HasPrefix(line, "#### ") {
		return "      - " + strings.TrimPrefix(line, "#### ")
	}

	// Handle bullet points
	if strings.HasPrefix(line, "- ") {
		return "  • " + strings.TrimPrefix(line, "- ")
	}
	if strings.HasPrefix(line, "* ") {
		return "  • " + strings.TrimPrefix(line, "* ")
	}

	// Handle numbered lists (simple)
	if matched, _ := regexp.MatchString(`^\d+\. `, line); matched {
		re := regexp.MustCompile(`^(\d+)\. (.*)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			return "  " + matches[1] + ". " + matches[2]
		}
	}

	// Handle inline markdown
	line = app.renderInlineMarkdown(line)

	return line
}

// renderInlineMarkdown processes inline markdown formatting
func (app *App) renderInlineMarkdown(text string) string {
	// Handle ***SUPER BOLD*** (triple asterisks) - keep as is for emphasis
	text = regexp.MustCompile(`\*\*\*([^*]+)\*\*\*`).ReplaceAllString(text, "🔥$1🔥")

	// Handle **bold** (double asterisks) - remove asterisks
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "$1")

	// Handle *italic* (single asterisks) - remove asterisks
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")

	// Handle __underlined__ - remove underscores but keep content
	text = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(text, "$1")

	// Handle ==highlighted== - remove equals but add visual indicator
	text = regexp.MustCompile(`==([^=]+)==`).ReplaceAllString(text, "✨$1✨")

	// Handle ^^large text^^ - remove carets but add visual indicator
	text = regexp.MustCompile(`\^\^([^^]+)\^\^`).ReplaceAllString(text, "📢$1📢")

	// Handle `code` - remove backticks but add visual indicator
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "[$1]")

	// Handle ~~strikethrough~~ - remove tildes
	text = regexp.MustCompile(`~~([^~]+)~~`).ReplaceAllString(text, "$1")

	return text
}
