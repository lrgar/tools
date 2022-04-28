package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5/plumbing"
)

type model struct {
	branches []*plumbing.Reference
	cursor   int
	selected bool
}

func main() {
	r, err := OpenRepo()
	if err != nil {
		log.Fatalf("Failed to open git repository: %v", err)
	}
	defer r.Close()

	branches, err := r.Branches()
	if err != nil {
		log.Fatalf("Failed to get branches: %v", err)
	}

	program := tea.NewProgram(model{branches: branches})
	retModel, err := program.StartReturningModel()
	if err != nil {
		log.Fatalf("Error while running UI: %v", err)
	}

	m := retModel.(model)
	if !m.selected {
		return
	}

	if err := r.SwitchBranch(m.branches[m.cursor]); err != nil {
		log.Fatal("Failed to switch branch: %v", err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up":
			m.cursor = (m.cursor + len(m.branches) - 1) % len(m.branches)

		case "down":
			m.cursor = (m.cursor + 1) % len(m.branches)

		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Select branch:\n"
	for i, branch := range m.branches {
		if m.cursor == i {
			s += fmt.Sprintf("\u001b[32m> %s\u001b[0m\n", branch.Name().Short())
		} else {
			s += fmt.Sprintf("  %s\n", branch.Name().Short())
		}
	}
	return s
}
