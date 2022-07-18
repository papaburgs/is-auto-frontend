package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/papaburgs/is-auto-frontend/pkg/stnetbox"
)

// var docStyle = lipgloss.NewStyle().Margin(1, 2)

const (
	columnKeyName   = "name"
	columnKeyStatus = "status"
	columnKeyAccess = "access"
	columnKeySync   = "sync"
)

const (
	checkmark = "✅"
	redex     = "❌"
)

type server struct {
	name   string
	status string
	access bool
	sync   bool
}

type model struct {
	tableModel table.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func NewModel(rows []table.Row) model {
	columns := []table.Column{
		table.NewColumn(columnKeyName, "Server Name", 20),
		table.NewColumn(columnKeyStatus, "Status", 16),
		table.NewColumn(columnKeyAccess, "Have Access", 5).WithStyle(
			lipgloss.NewStyle().Align(lipgloss.Center)),
		table.NewColumn(columnKeySync, "Sync", 5).WithStyle(
			lipgloss.NewStyle().Align(lipgloss.Center)),
	}

	model := model{
		// Throw features in... the point is not to look good, it's just reference!
		tableModel: table.New(columns).
			WithRows(rows).
			SelectableRows(true).
			Focused(true).
			WithStaticFooter("Footer!").
			WithPageSize(15).
			WithSelectedText(" ", "✓").
			WithBaseStyle(
				lipgloss.NewStyle().
					BorderForeground(lipgloss.Color("#a38")).
					Foreground(lipgloss.Color("#a7a")).
					Align(lipgloss.Left),
			).
			SortByAsc(columnKeyName),
	}

	model.updateFooter()

	return model
}
func (m *model) updateFooter() {
	highlightedRow := m.tableModel.HighlightedRow()

	footerText := fmt.Sprintf(
		"Pg. %d/%d - Currently looking at ID: %s",
		m.tableModel.CurrentPage(),
		m.tableModel.MaxPages(),
		highlightedRow.Data[columnKeyName],
	)

	m.tableModel = m.tableModel.WithStaticFooter(footerText)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.tableModel, cmd = m.tableModel.Update(msg)
	cmds = append(cmds, cmd)

	// We control the footer text, so make sure to update it
	m.updateFooter()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		}

	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	body := strings.Builder{}

	body.WriteString("Press space/enter to select a row, q or ctrl+c to quit\n")

	selectedIDs := []string{}

	for _, row := range m.tableModel.SelectedRows() {
		// Slightly dangerous type assumption but fine for demo
		selectedIDs = append(selectedIDs, row.Data[columnKeyName].(string))
	}

	body.WriteString(fmt.Sprintf("SelectedIDs: %s\n", strings.Join(selectedIDs, ", ")))

	body.WriteString(m.tableModel.View())

	body.WriteString("\n")

	return body.String()
}

func main() {
	s, _ := stnetbox.GetInv()
	rows := []table.Row{}
	for i, item := range s {
		thisServer := table.NewRow(table.RowData{
			columnKeyName:   item.Name,
			columnKeyStatus: "unknown",
			columnKeyAccess: redex,
			columnKeySync:   checkmark,
		})
		rows = append(rows, thisServer)
		if i > 16 {
			break
		}
	}

	p := tea.NewProgram(NewModel(rows), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
