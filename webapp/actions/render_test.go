package actions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownToHTML_Empty(t *testing.T) {
	assert.Equal(t, "", string(markdownToHTML("")))
}

func TestMarkdownToHTML_Bold(t *testing.T) {
	out := string(markdownToHTML("**gala**"))
	assert.Contains(t, out, "<strong>gala</strong>")
}

func TestMarkdownToHTML_Italic(t *testing.T) {
	out := string(markdownToHTML("_italique_"))
	assert.Contains(t, out, "<em>italique</em>")
}

func TestMarkdownToHTML_XSS_ScriptTag(t *testing.T) {
	out := string(markdownToHTML("<script>alert(1)</script>"))
	assert.False(t, strings.Contains(out, "<script>"), "raw <script> must be stripped")
}

func TestMarkdownToHTML_XSS_EventHandler(t *testing.T) {
	out := string(markdownToHTML("<img src=x onerror=alert(1)>"))
	assert.False(t, strings.Contains(out, "onerror"), "inline event handler must be stripped")
}
