package board

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-homedir"
	"kancli/lib/task"
	"math/rand"
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

const (
	BoardView state = iota
	TaskView
	TaskEdit
)

type Board struct {
	Lists   []list.Model
	Pager   viewport.Model
	Config  Config
	Focused task.Status
	Ready   bool
	Width   int
	Height  int
	Editor  textarea.Model
	state   state
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

func configPath() string {
	p, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configFile := path.Join(p, "kancli", "config.json")
	return configFile
}

func getConfig() Config {
	configFile := configPath()
	p, err := os.UserHomeDir()
	err = os.MkdirAll(path.Join(p, "kancli", "tasks"), os.ModePerm)
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

func (b *Board) Move(from, to task.Status) tea.Msg {
	currentStatus := from
	if len(b.Lists[from].Items()) == 0 {
		return nil
	}
	i := b.CurrentList().Index()
	nextStatus := to
	items := b.CurrentList().Items()
	item := items[i]
	items = removeItemFromSlice(items, i)
	nextItems := b.Lists[nextStatus].Items()
	nextItems = append(nextItems, item)
	msg := UpdateLists{Lists: map[task.Status][]list.Item{
		currentStatus: items,
		nextStatus:    nextItems,
	}}
	return msg
}

func (b *Board) MoveToPrev() tea.Msg {
	return b.Move(b.Focused, task.Prev(b.Focused))
}

func (b *Board) MoveToNext() tea.Msg {
	return b.Move(b.Focused, task.Next(b.Focused))
}

func (b Board) CurrentTask() task.Task {
	return b.CurrentList().SelectedItem().(task.Task)
}

func (b Board) UpdateCurrentTask(newContent string) tea.Cmd {
	return func() tea.Msg {
		i := b.Lists[b.Focused].Index()
		items := b.Lists[b.Focused].Items()
		t := items[i].(task.Task)
		t.Content = newContent
		if _, err := os.Stat(t.Path); !errors.Is(err, os.ErrNotExist) && err != nil {
			panic(err)
		}
		err := os.WriteFile(t.Path, []byte(newContent), 0666)
		if err != nil {
			panic(err)
		}
		items[i] = list.Item(t)
		return UpdateLists{Lists: map[task.Status][]list.Item{
			b.Focused: items,
		}}
	}
}

func (b Board) Quit() tea.Cmd {
	configStr, err := json.Marshal(b.Config)
	if err != nil {
		return tea.Quit
	}
	_ = os.WriteFile(configPath(), configStr, 0666)
	return tea.Quit
}

func (b Board) DeleteCurrent() tea.Cmd {
	return func() tea.Msg {
		i := b.Lists[b.Focused].Index()
		items := b.Lists[b.Focused].Items()
		return UpdateLists{Lists: map[task.Status][]list.Item{
			b.Focused: removeItemFromSlice(items, i),
		}}
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (b Board) CreateTask() tea.Cmd {
	updateListCmd := func() tea.Msg {
		h, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		p := path.Join(h, "kancli", "tasks", RandStringBytes(10)+".md")
		err = os.WriteFile(p, []byte("\n"), 0666)
		if err != nil {
			panic(err)
		}
		t, err := task.NewFromFile(p)
		if err != nil {
			panic(err)
		}
		l := append(b.CurrentList().Items(), list.Item(t))
		return UpdateLists{Lists: map[task.Status][]list.Item{
			b.Focused: l,
		}}
	}
	editCmd := func() tea.Msg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e"), Alt: false}
	}
	return tea.Sequence(updateListCmd, editCmd)
}

func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	//fmt.Printf("%+v\n", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if b.state == TaskEdit {
			if msg.Type == tea.KeyEsc {
				b.state = BoardView
				return b, b.UpdateCurrentTask(b.Editor.Value())
			}
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			if b.state == TaskView {
				b.state = BoardView
				return b, nil
			}
			return b, b.Quit()
		case "right":
			b.Focused = task.Next(b.Focused)
		case "left":
			b.Focused = task.Prev(b.Focused)
		case " ", "ctrl+right":
			return b, b.MoveToNext
		case "ctrl+left":
			return b, b.MoveToPrev
		case "enter":
			b.state = TaskView
			b.Pager.Height = b.Height
			return b, b.CurrentTask().Render(b.Width)
		case "delete":
			return b, b.DeleteCurrent()
		case "n":
			return b, b.CreateTask()
		case "e":
			b.state = TaskEdit
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
			b.Config.Lists[i].Items = []string{}
			for _, vv := range v {
				b.Config.Lists[i].Items = append(b.Config.Lists[i].Items, vv.(task.Task).Path)
			}
		}
	case tea.WindowSizeMsg:
		b.Width = msg.Width
		b.Height = msg.Height
		b.Pager.Width = b.Width
		b.Pager.Height = b.Height
		b.Editor.SetHeight(msg.Height)
		b.Editor.SetWidth(msg.Width)
		if b.state == TaskView {
			cmd = b.CurrentTask().Render(b.Width)
		}
	case task.ContentRenderedMsg:
		if b.state == TaskView {
			b.Pager.SetContent(string(msg))
		} else if b.state == TaskEdit {
			b.Editor.SetValue(b.CurrentTask().Content)
			for b.Editor.Line() > 0 {
				b.Editor.CursorUp()
			}
			b.Editor.CursorStart()
			return b, b.Editor.Focus()
		}
	}
	if b.state == TaskView {
		b.Pager, cmd = b.Pager.Update(msg)
		return b, cmd
	} else if b.state == TaskEdit {
		fmt.Printf("%+v", msg)
		b.Editor, cmd = b.Editor.Update(msg)
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
	if b.state == TaskView {
		return b.Pager.View()
	}
	if b.state == TaskEdit {
		return b.Editor.View()
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
	ta := textarea.New()
	ta.CharLimit = 0
	return Board{Editor: ta}
}
