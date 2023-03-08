package kancli

import (
	tea "github.com/charmbracelet/bubbletea"
	"kancli/lib/board"
)

type CurrentModel int

const (
	Board CurrentModel = iota
	Task
)

type RootModel struct {
	Models   []tea.Model
	Current  CurrentModel
	Quitting bool
}

func (r RootModel) Init() tea.Cmd {
	var inits []tea.Cmd
	for _, v := range r.Models {
		inits = append(inits, v.Init())
	}
	return tea.Batch(inits...)
}

func (r RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			r.Quitting = true
			return r, tea.Quit
		}
	}
	return r.Models[r.Current].Update(msg)
}

func (r RootModel) View() string {
	if r.Quitting {
		return ""
	}
	return r.Models[r.Current].View()
}

func Run() error {
	rootModel := RootModel{Models: []tea.Model{board.New()}, Current: Board}
	p := tea.NewProgram(rootModel)
	_, err := p.Run()
	if err != nil {
		return err
	}
	return nil
}
