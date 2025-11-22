package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "5", Dark: "5"}).
			Padding(0, 1)

	counterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}).
			Background(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}).
			Padding(0, 0)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}).
			Foreground(lipgloss.AdaptiveColor{Light: "7", Dark: "7"})

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Action styles with semantic colors
	acceptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "2", Dark: "2"}). // Green
			Bold(true)

	rejectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "1"}). // Red
			Bold(true)

	skipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "3", Dark: "3"}). // Yellow
			Bold(true)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"})

	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"})
)

type model struct {
	snapshots    []string
	current      int
	newSnap      *files.Snapshot
	accepted     *files.Snapshot
	diffLines    []diff.DiffLine
	choice       string
	done         bool
	err          error
	acceptedAll  int
	rejectedAll  int
	skippedAll   int
	actionResult string
	viewport     viewport.Model
	ready        bool
	width        int
	height       int
}

func initialModel() (model, error) {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return model{}, err
	}

	if len(snapshots) == 0 {
		return model{done: true}, nil
	}

	m := model{
		snapshots: snapshots,
		current:   0,
	}

	if err := m.loadCurrentSnapshot(); err != nil {
		return model{}, err
	}

	return m, nil
}

func (m *model) loadCurrentSnapshot() error {
	if m.current >= len(m.snapshots) {
		m.done = true
		return nil
	}

	testName := m.snapshots[m.current]

	newSnap, err := files.ReadSnapshot(testName, "new")
	if err != nil {
		return err
	}
	m.newSnap = newSnap

	accepted, err := files.ReadSnapshot(testName, "accepted")
	if err == nil {
		m.accepted = accepted
		diffLines := computeDiffLines(accepted, newSnap)
		m.diffLines = diffLines
	} else {
		m.accepted = nil
		m.diffLines = nil
	}

	return nil
}

func computeDiffLines(old, new *files.Snapshot) []diff.DiffLine {
	return diff.Histogram(old.Content, new.Content)
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 1
		footerHeight := 1
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.updateViewportContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit

		case "a":
			// Accept current snapshot
			testName := m.snapshots[m.current]
			if err := files.AcceptSnapshot(testName); err != nil {
				m.err = err
			} else {
				m.acceptedAll++
				m.current++
				if err := m.loadCurrentSnapshot(); err != nil {
					m.err = err
				}
				if m.done {
					return m, tea.Quit
				}
				m.updateViewportContent()
			}

		case "r":
			// Reject current snapshot
			testName := m.snapshots[m.current]
			if err := files.RejectSnapshot(testName); err != nil {
				m.err = err
			} else {
				m.rejectedAll++
				m.current++
				if err := m.loadCurrentSnapshot(); err != nil {
					m.err = err
				}
				if m.done {
					return m, tea.Quit
				}
				m.updateViewportContent()
			}

		case "s":
			// Skip current snapshot
			m.skippedAll++
			m.current++
			if err := m.loadCurrentSnapshot(); err != nil {
				m.err = err
			}
			if m.done {
				return m, tea.Quit
			}
			m.updateViewportContent()

		case "A":
			// Accept all remaining
			for i := m.current; i < len(m.snapshots); i++ {
				if err := files.AcceptSnapshot(m.snapshots[i]); err != nil {
					m.err = err
					break
				}
				m.acceptedAll++
			}
			m.done = true
			return m, tea.Quit

		case "R":
			// Reject all remaining
			for i := m.current; i < len(m.snapshots); i++ {
				if err := files.RejectSnapshot(m.snapshots[i]); err != nil {
					m.err = err
					break
				}
				m.rejectedAll++
			}
			m.done = true
			return m, tea.Quit

		case "S":
			// Skip all remaining
			m.skippedAll = len(m.snapshots) - m.current
			m.done = true
			return m, tea.Quit
		}
	}

	// Handle viewport scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updateViewportContent() {
	if !m.ready {
		return
	}

	var b strings.Builder

	// Show diff or new snapshot
	if m.accepted != nil && m.diffLines != nil {
		b.WriteString(pretty.DiffSnapshotBox(m.accepted, m.newSnap, m.diffLines))
	} else {
		if m.newSnap != nil {
			b.WriteString(pretty.NewSnapshotBox(m.newSnap))
		}
	}

	// Add action options below the snapshot/diff box
	b.WriteString("\n")
	acceptLine := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("[a]"),
		helpTextStyle.Render(" "),
		acceptStyle.Render("accept"),
	)
	b.WriteString(acceptLine)
	b.WriteString("\n")

	rejectLine := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("[r]"),
		helpTextStyle.Render(" "),
		rejectStyle.Render("reject"),
	)
	b.WriteString(rejectLine)
	b.WriteString("\n")

	skipLine := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("[s]"),
		helpTextStyle.Render(" "),
		skipStyle.Render("skip"),
	)
	b.WriteString(skipLine)

	m.viewport.SetContent(contentStyle.Render(b.String()))
	m.viewport.GotoTop()
}

func (m model) View() string {
	if m.done {
		if len(m.snapshots) == 0 {
			return pretty.Success("✓ No new snapshots to review\n")
		}

		// Build summary from counts
		var summary []string
		if m.acceptedAll > 0 {
			summary = append(summary, fmt.Sprintf("✓ Accepted %d", m.acceptedAll))
		}
		if m.rejectedAll > 0 {
			summary = append(summary, fmt.Sprintf("⊘ Rejected %d", m.rejectedAll))
		}
		if m.skippedAll > 0 {
			summary = append(summary, fmt.Sprintf("⊘ Skipped %d", m.skippedAll))
		}

		if len(summary) > 0 {
			return pretty.Success(strings.Join(summary, " • ") + "\n")
		}
		return pretty.Success("\n✓ Review complete\n")
	}

	if m.err != nil {
		return pretty.Error("Error: " + m.err.Error() + "\n")
	}

	if !m.ready {
		return "\n  Initializing..."
	}

	// Header
	snapshotTitle := m.snapshots[m.current] // fallback to test name
	if m.newSnap != nil && m.newSnap.Title != "" {
		snapshotTitle = m.newSnap.Title
	}
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render("Review Snapshots"),
		counterStyle.Render(fmt.Sprintf("[%d/%d] %s", m.current+1, len(m.snapshots), snapshotTitle)),
	)
	headerStyled := statusBarStyle.Width(m.width).Render(header)

	// Footer with snapshot filename and scroll info
	snapshotFile := files.SnapshotFileName(m.snapshots[m.current]) + ".snap.new"
	fileInfo := helpStyle.Render(snapshotFile)
	scrollInfo := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	scrollStyled := helpStyle.Render(scrollInfo)

	// Calculate spacing between filename and scroll percentage
	totalFooterWidth := lipgloss.Width(fileInfo) + lipgloss.Width(scrollStyled)
	spacing := max(m.width-totalFooterWidth-2, 1)

	// Create footer with filename on left and scroll info on right
	footer := lipgloss.JoinHorizontal(lipgloss.Bottom,
		fileInfo,
		strings.Repeat(" ", spacing),
		scrollStyled,
	)
	footerStyled := statusBarStyle.Width(m.width).Render(footer)

	// Viewport content
	// TODO: it would be nice if we could show the input on the right side?
	// - (probably optionally, with a keybind -- hidden by default)
	viewportContent := m.viewport.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyled,
		viewportContent,
		footerStyled,
	)
}

func acceptAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.AcceptSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Success("✓ Accepted %d snapshot(s)\n"), len(snapshots))
	return nil
}

func rejectAll() error {
	snapshots, err := files.ListNewSnapshots()
	if err != nil {
		return err
	}

	for _, testName := range snapshots {
		if err := files.RejectSnapshot(testName); err != nil {
			return err
		}
	}

	fmt.Printf(pretty.Warning("⊘ Rejected %d snapshot(s)\n"), len(snapshots))
	return nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "accept-all":
			if err := acceptAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "reject-all":
			if err := rejectAll(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			fmt.Println(`Usage: shutter-tui [COMMAND]

Commands:
  review      Review and accept/reject new snapshots (default)
  accept-all  Accept all new snapshots
  reject-all  Reject all new snapshots
  help        Show this help message

Interactive Controls:
  a           Accept current snapshot
  r           Reject current snapshot
  s           Skip current snapshot
  A           Accept all remaining snapshots
  R           Reject all remaining snapshots
  S           Skip all remaining snapshots
  q           Quit`)
			return
		}
	}

	m, err := initialModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m.done && len(m.snapshots) == 0 {
		fmt.Println(m.View())
		return
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
