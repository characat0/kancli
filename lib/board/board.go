package board

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"kancli/lib/task"
	"os"
	"path"
	"strings"
)

type Config struct {
	Lists []struct {
		Items []string `json:"items"`
	} `json:"Lists"`
}

type state int

type Board struct {
	Lists      []list.Model
	Pager      viewport.Model
	Config     Config
	Focused    task.Status
	Ready      bool
	Width      int
	Height     int
	RenderTask bool
	state      state
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
	listString := make([]string, task.NumberOfStatus)
	for i := range listString {
		listString[i] = `{ "items": [] }`
	}
	return []byte(fmt.Sprintf(`{
	"Lists": [
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
	for i, v := range task.StatusNames {
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
	Pager  viewport.Model
}

func (b Board) Init() tea.Cmd {
	return func() tea.Msg {
		lists := make([]list.Model, task.NumberOfStatus)
		config := getConfig()
		lists = InitLists(config, lists)
		pager := viewport.New(0, 0)
		pager.YPosition = 0
		return ReadyMsg{Lists: lists, Config: config, Pager: pager}
	}
}

func removeItemFromSlice[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

type UpdateLists struct {
	Lists map[task.Status][]list.Item
}

func (b Board) CurrentList() list.Model {
	return b.Lists[b.Focused]
}

func (b *Board) MoveToNext() tea.Msg {
	if len(b.Lists[b.Focused].Items()) == 0 {
		return nil
	}
	i := b.CurrentList().Index()
	currentStatus := b.Focused
	nextStatus := task.Next(currentStatus)
	items := b.CurrentList().Items()
	item := items[i]
	items = removeItemFromSlice(items, i)
	nextItems := b.Lists[nextStatus].Items()
	nextItems = append(nextItems, item)
	msg := UpdateLists{Lists: map[task.Status][]list.Item{
		currentStatus: items,
		nextStatus:    nextItems,
	}}
	selectedTask := item.(task.Task)
	b.Config.Lists[b.Focused].Items = removeItemFromSlice(b.Config.Lists[b.Focused].Items, i)
	b.Config.Lists[nextStatus].Items = append(b.Config.Lists[nextStatus].Items, selectedTask.Path)
	return msg
}

func (b Board) CurrentTask() task.Task {
	return b.CurrentList().SelectedItem().(task.Task)
}

func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return b, tea.Quit
		case "right":
			b.Focused = task.Next(b.Focused)
		case "left":
			b.Focused = task.Prev(b.Focused)
		case " ":
			return b, b.MoveToNext
		case "enter":
			b.RenderTask = !b.RenderTask
			b.Pager.Height = b.Height
			return b, b.CurrentTask().Render(b.Width)
		}
	case ReadyMsg:
		b.Ready = true
		b.Lists = msg.Lists
		b.Config = msg.Config
		b.Pager = msg.Pager

	case UpdateLists:
		for i, v := range msg.Lists {
			b.Lists[i].SetItems(v)
		}
	case tea.WindowSizeMsg:
		b.Width = msg.Width
		b.Height = msg.Height
		b.Pager.Width = b.Width
		b.Pager.Height = b.Height
		if b.RenderTask {
			return b, b.CurrentTask().Render(b.Width)
		}
	case task.ContentRenderedMsg:
		b.Pager.SetContent(string(msg))
	}
	if b.RenderTask {
		b.Pager, cmd = b.Pager.Update(msg)
		return b, cmd
	} else if b.Ready {
		b.Lists[b.Focused], cmd = b.Lists[b.Focused].Update(msg)
	}
	return b, tea.Batch(cmds...)
}

func (b Board) View() string {
	if !b.Ready {
		return fmt.Sprintf("loading... [%+v], %v, %v", b.Lists, b.Width, b.Height)
	}
	if b.RenderTask {
		return b.Pager.View()
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
