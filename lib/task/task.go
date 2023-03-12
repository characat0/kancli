package task

import (
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
	stringParts := strings.SplitN(strings.ReplaceAll(t.Content, "\r\n", "\n"), "\n", 2)
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
	stringParts := strings.SplitN(strings.ReplaceAll(t.Content, "\r\n", "\n"), "\n", 2)
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
