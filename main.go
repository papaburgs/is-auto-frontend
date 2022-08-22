package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/evertras/bubble-table/table"
	"github.com/gliderlabs/ssh"
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
	tableModel    table.Model
	term          string
	width         int
	height        int
	authenticated bool
}

func (m model) Init() tea.Cmd {
	log.Println("start init")
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
	default:
		log.Printf("%s", msg)

	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	body := strings.Builder{}

	if m.width < 120 {
		return "Width needs to be greater than 120"
	}

	if m.authenticated {
		body.WriteString("Valid auth\n")
	} else {
		body.WriteString("READ ONLY\n")
	}

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

// authenticate will read the files in a directory
//  - which can be populated in a config map
// each file is a username, contents are the pub key
// if the pub key matches the one that we authenticated with and the filename
// matches the user that signed in, we are authenticated, if not, result if false
func authenticate(s ssh.Session) bool {
	/*
		files, err := ioutil.ReadDir("./keys")
		if err != nil {
			log.Println("Could not read key directory")
			log.Println(err)
			return false
		}
		pubKeyFilename := fmt.Sprintf("%s.pub", s.User())
		log.Println(s.PublicKey())
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			log.Println(file.Name())
			if file.Name() == pubKeyFilename {
				log.Println("found a match")
				contents, err := os.ReadFile(fmt.Sprintf("./keys/%s", file.Name()))
				if err != nil {
					log.Printf("Could not read %s - %s", file.Name(), err)
					continue
				}
				log.Println(contents)
				thisPublicKey, _, _, _, err := ssh.ParseAuthorizedKey(contents)
				if err != nil {
					log.Printf("Could not parse key from %s - %s", file.Name(), err)
					continue
				}
				if ssh.KeysEqual(s.PublicKey(), thisPublicKey) {
					log.Println("match")
					return true
				}
			}
		}
		return false
	*/
	keys := []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOZ/NkruEpZxh+o/dUuQaR7NOzgFJCYcXnkKOJwtL4b3 olsod2@slc1053333.stholdco.com",
	}
	for keyIndex, key := range keys {
		thisPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(key)))
		if err != nil {
			log.Printf("Could not parse key from at index %v", keyIndex)
			continue
		}
		if ssh.KeysEqual(s.PublicKey(), thisPublicKey) {
			log.Println("match")
			return true
		}
	}
	return false
}

const (
	host = "0.0.0.0"
	port = 39722
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	pty, _, active := s.Pty()
	if !active {
		wish.Fatalln(s, "no active terminal, skipping")
		return nil, nil
	}
	servers, _ := stnetbox.GetInv()
	rows := []table.Row{}
	for i, item := range servers {
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

	m := NewModel(rows)
	m.term = pty.Term
	m.width = pty.Window.Width
	m.height = pty.Window.Height
	m.authenticated = authenticate(s)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool { return true }),
		wish.WithMiddleware(
			bm.Middleware(teaHandler),
			lm.Middleware(),
			func(h ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					carlos, _, _, _, _ := ssh.ParseAuthorizedKey(
						[]byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOZ/NkruEpZxh+o/dUuQaR7NOzgFJCYcXnkKOJwtL4b3"),
					)
					switch {
					case ssh.KeysEqual(s.PublicKey(), carlos):
						wish.Println(s, "Hey Carlos!")
					default:
						wish.Println(s, "Hey, I don't know who you are!")
						log.Println("don't you see this?")
					}
					h(s)
				}
			},
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
