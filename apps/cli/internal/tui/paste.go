package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const (
	pasteCollapseLines = 10
	pasteCollapseChars = 1000
)

var pasteToken = regexp.MustCompile(`\[Pasted text #(\d+) \+\d+ lines\]`)

// pastes holds the long pastes the prompt shows as placeholders, numbered
// from 1, and the prompt value right after the last one went in, so a paste
// the terminal split into chunks merges back into one placeholder.
type pastes struct {
	texts []string
	after string
}

// insertPaste puts pasted text at the cursor. A long paste becomes a
// "[Pasted text #N +X lines]" placeholder that expandPastes swaps back for
// the real text on submit.
func (m *Model) insertPaste(content string) tea.Cmd {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if m.prompt.Value() == "" {
		m.pasted = pastes{}
	}
	switch {
	case m.mergesWithLastPaste():
		m.growLastPaste(content)
	case strings.Count(content, "\n") >= pasteCollapseLines || len(content) > pasteCollapseChars:
		m.pasted.texts = append(m.pasted.texts, content)
		m.prompt.InsertString(placeholder(len(m.pasted.texts), content))
		m.pasted.after = m.prompt.Value()
	default:
		m.insertLines(content)
	}
	m.refreshMenu()
	return nil
}

func (m *Model) mergesWithLastPaste() bool {
	return len(m.pasted.texts) > 0 && m.pasted.after != "" && m.prompt.Value() == m.pasted.after
}

func (m *Model) growLastPaste(content string) {
	n := len(m.pasted.texts)
	old := placeholder(n, m.pasted.texts[n-1])
	m.pasted.texts[n-1] += content
	value := strings.Replace(m.prompt.Value(), old, placeholder(n, m.pasted.texts[n-1]), 1)
	m.prompt.SetValue(value)
	m.prompt.CursorEnd()
	m.pasted.after = m.prompt.Value()
}

func (m *Model) insertLines(content string) {
	for i, line := range strings.Split(content, "\n") {
		if i > 0 {
			m.prompt, _ = m.prompt.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModAlt})
		}
		if line != "" {
			m.prompt.InsertString(line)
		}
	}
}

func placeholder(n int, text string) string {
	return fmt.Sprintf("[Pasted text #%d +%d lines]", n, strings.Count(text, "\n")+1)
}

// takePrompt returns the prompt's text with every placeholder expanded, and
// forgets the pastes: what leaves the prompt is always the full text.
func (m *Model) takePrompt() string {
	text := pasteToken.ReplaceAllStringFunc(m.prompt.Value(), func(token string) string {
		n, err := strconv.Atoi(pasteToken.FindStringSubmatch(token)[1])
		if err != nil || n < 1 || n > len(m.pasted.texts) {
			return token
		}
		return m.pasted.texts[n-1]
	})
	m.pasted = pastes{}
	return text
}

// echo is a question as the transcript shows it: a long one keeps its first
// lines and says how many more went to the model.
func echo(question string) string {
	lines := strings.Split(question, "\n")
	if len(lines) <= pasteCollapseLines {
		return question
	}
	return strings.Join(lines[:3], "\n") + fmt.Sprintf("\n… +%d lines", len(lines)-3)
}
