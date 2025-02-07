package task

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/mitchellh/go-homedir"
	"os"
	"strings"
)

type Task struct {
	Path    string
	Content string
}

func NewFromFile(path string) (Task, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return Task{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Task{}, err
	}
	return Task{Content: string(content), Path: path}, nil
}

func (t Task) FilterValue() string {
	return t.Content
}

func (t Task) Title() string {
	c := t.Content
	c = strings.TrimSpace(c)
	stringParts := strings.SplitN(strings.ReplaceAll(c, "\r\n", "\n"), "\n", 2)
	title := ""
	if len(stringParts) >= 1 {
		title = stringParts[0]
	}
	if strings.HasPrefix(title, "# ") {
		title = title[2:]
	}
	if title == "" {
		title = "Untitled"
	}
	return title
}

func (t Task) Description() string {
	c := t.Content
	c = strings.TrimSpace(c)
	stringParts := strings.SplitN(strings.ReplaceAll(c, "\r\n", "\n"), "\n", 2)
	description := ""

	if len(stringParts) >= 2 {
		description = strings.TrimSpace(stringParts[1])
	}
	if description == "" {
		description = "No Description"
	}
	description = strings.ReplaceAll(description, "\n", "↵ ")
	return description
}

type ContentRenderedMsg string

func (t Task) Render(width int) tea.Cmd {
	return func() tea.Msg {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
			glamour.WithPreservedNewLines(),
			glamour.WithEmoji(),
		)
		if err != nil {
			panic(err)
		}
		out, err := r.Render(t.Content)
		if err != nil {
			panic(err)
		}
		// trim lines
		lines := strings.Split(out, "\n")

		var content string
		for i, s := range lines {
			content += strings.TrimSpace(s)

			// don't add an artificial newline after the last split
			if i+1 < len(lines) {
				content += "\n"
			}
		}
		return ContentRenderedMsg(content)
	}
}
