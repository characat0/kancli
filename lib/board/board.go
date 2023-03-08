package board

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"kancli/lib/status"
	"kancli/lib/task"
	"os"
	"path"
	"strings"
)

type Config struct {
	Lists []struct {
		Items []string `json:"items"`
	} `json:"lists"`
}

type Board struct {
	Lists   []list.Model
	Config  Config
	Focused status.Status
	Ready   bool
	Width   int
	Height  int
}

var (
	columnStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.HiddenBorder())
	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

const (
	widthDivisor = 3
)

func (b Board) Resize() {
	width := b.Width
	height := b.Height
	columnStyle.Width(width / widthDivisor)
	focusedStyle.Width(width / widthDivisor)
	columnStyle.Height(height - widthDivisor)
	focusedStyle.Height(height - widthDivisor)
	for i, l := range b.Lists {
		var x, y int
		if i == int(b.Focused) {
			x, y = focusedStyle.GetFrameSize()
		} else {
			x, y = columnStyle.GetFrameSize()
		}
		l.SetSize(width/widthDivisor-x, height-y)
		b.Lists[i] = l
	}

}

func defaultConfig() []byte {
	listString := make([]string, status.NumberOfStatus)
	for i := range listString {
		listString[i] = `{ "items": [] }`
	}
	return []byte(fmt.Sprintf(`{
	"lists": [
%s
	]
}`, strings.Join(listString, ",\n")))
}

func getConfig() Config {
	p, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configFile := path.Join(p, "kancli", "config.json")
	err = os.MkdirAll(path.Join(p, "kancli"), os.ModePerm)
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(configFile, defaultConfig(), 0666)
		if err != nil {
			panic(err)
		}
	}
	var config Config
	f, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(f, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func InitLists(config Config, lists []list.Model) []list.Model {
	defaultList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	defaultList.SetShowHelp(false)
	for i, v := range status.StatusNames {
		lists[i] = defaultList
		lists[i].Title = v
		var items []list.Item
		for _, filePath := range config.Lists[i].Items {
			t, err := task.NewFromFile(filePath)
			if err != nil {
				panic(err)
			}
			items = append(items, t)
		}
		lists[i].SetItems(items)
	}
	return lists
}

type ReadyMsg struct {
	Lists  []list.Model
	Config Config
}

func (b Board) Init() tea.Cmd {
	return func() tea.Msg {
		lists := make([]list.Model, status.NumberOfStatus)
		config := getConfig()
		lists = InitLists(config, lists)
		return ReadyMsg{Lists: lists, Config: config}
	}
}

func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "right":
			b.Focused = status.Next(b.Focused)
		case "left":
			b.Focused = status.Prev(b.Focused)
		}
	case ReadyMsg:
		b.Ready = true
		b.Lists = msg.Lists
		b.Config = msg.Config

	case tea.WindowSizeMsg:
		b.Width = msg.Width
		b.Height = msg.Height
	}
	var cmd tea.Cmd
	if b.Ready {
		b.Lists[b.Focused], cmd = b.Lists[b.Focused].Update(msg)
	}
	return b, cmd
}

func (b Board) View() string {
	if !b.Ready {
		return fmt.Sprintf("loading... [%+v], %v, %v", b.Lists, b.Width, b.Height)
	}
	b.Resize()
	var views []string
	for i, v := range b.Lists {
		if i == int(b.Focused) {
			views = append(views, focusedStyle.Render(v.View()))
		} else {
			views = append(views, columnStyle.Render(v.View()))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, views...)
}

func New() Board {
	return Board{}
}
