package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("111"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	menuMainColor     = "205"
	menuSettingsColor = "111"
	// menuSMTPcolor       = "184"
	textPromptColor = "141" //"100" //nice: 141
	// textInputColor  = "193" //"40" //nice: 193
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textResultJob       = "141" //PINK"205"
	textJobOutcomeFront = "216"
	// txtJobOutcomeBack   = "205"

	menuTOP = []string{
		"Toggle Provider",
		"Change Number of Boxes to deploy",
		"Enter Digital Ocean API Token",
		"Enter AWS Key",
		"Enter AWS Secret",
		"RUN Box Deployments",
		"DELETE all Boxes",
		"Save Settings",
	}
	doAPI = "xx"
)

// App States
type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateResultDisplay
	StateSpinner
	StateTextInput
)

// Messsage returend when the background job finishes
type backgroundJobMsg struct {
	result string
}

type JobList int

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	headerIP            string
	state               MenuState
	prevState           MenuState
	prevMenuState       MenuState
	spinner             spinner.Model
	spinnerMsg          string
	backgroundJobResult string
	textInput           textinput.Model
	inputPrompt         string
	textInputError      bool
	jobOutcome          string
	manifestProvider    string
	manifestBoxes       int
	manifestLinodeAPI   string
	manifestAWSkey      string
	manifestAWSsecret   string
}

func (m MenuList) Init() tea.Cmd {
	return nil
}

func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.updateMainMenu(msg)
	case StateSpinner:
		return m.updateSpinner(msg)
	case StateTextInput:
		return m.updateTextInput(msg)
	case StateResultDisplay:
		return m.updateResultDisplay(msg)
	default:
		return m, nil
	}
}

func (m *MenuList) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			err := clipboard.WriteAll(m.headerIP)
			if err != nil {
				fmt.Println("Failed to copy to clipboard:", err)
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				switch m.choice {
				case menuTOP[0]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundPEPA())
				case menuTOP[5]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobCreateBox())
				case menuTOP[6]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobDeleteBox())
				default: //all other options go to startBackgroundJob
					m.prevState = m.state
					m.list.Title = "Main Menu->Settings"
					m.header = "Pepa Header"
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
					m.state = StateMainMenu
					m.updateListItems()
					return m, nil
				}
			}
			return m, nil
		}
		// case jobListMsg:

		// 	// m.state = StateResultDisplay
		// 	// return m, nil
		// 	m.prevState = m.state
		// 	m.state = StateSpinner
		// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.textInputError = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.inputPrompt = m.textInput.Value() // User pressed enter, save the input

			switch m.inputPrompt {
			case m.inputPrompt:
				fmt.Println("yep")
			}
			m.prevState = m.state
			m.state = StateResultDisplay
			return m, nil
		case tea.KeyEsc:
			// m.state = StateSettingsMenu
			m.state = m.prevState
			return m, nil
		}
	}

	return m, cmd
}

func (m *MenuList) updateSpinner(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "q", "esc":
		// 	m.backgroundJobResult = "Job Cancelled"
		// 	m.state = StateResultDisplay
		// 	return m, nil
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		m.backgroundJobResult = m.jobOutcome + "\n\n" + msg.result + "\n"
		m.state = StateResultDisplay
		return m, nil
	// case continueJobs:
	// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.textInputError {
				m.state = m.prevState
			} else {
				m.state = m.prevMenuState
			}
			m.updateListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuList) viewResultDisplay() string {
	outro := "Press 'esc' to return."
	outroRender := lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true).Render(outro)
	lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true)
	if m.textInputError {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult)
	} else {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(m.backgroundJobResult)
	}
	return fmt.Sprintf("\n\n%s\n\n%s", m.backgroundJobResult, outroRender)

	// //repeat interval
	// if m.configSettings.Interval > 0 {

	// }
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu:
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return m.viewSpinner()
	case StateTextInput:
		return m.viewTextInput()
	case StateResultDisplay:
		return m.viewResultDisplay()
	default:
		return "Unknown state"
	}
}

func (m MenuList) viewSpinner() string {
	// tea.ClearScreen()
	spinnerBase := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.spinnerMsg)

	// return spinnerBase + m.jobOutcome
	return spinnerBase + lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome)
}

func (m MenuList) viewTextInput() string {
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor)).Bold(true)
	return fmt.Sprintf("\n\n%s\n\n%s", promptStyle.Render(m.inputPrompt), m.textInput.View())

}

func (m *MenuList) updateListItems() {
	switch m.state {
	case StateMainMenu:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, item(value))
		}
		m.list.SetItems(items)
		// case StateSettingsMenu:
		// 	items := []list.Item{}
		// 	for _, value := range menuSettings {
		// 		items = append(items, item(value[0]))
		// 	}
		// 	m.list.SetItems(items)
	}

	m.list.ResetSelected()
}

func (m *MenuList) backgroundPEPA() tea.Cmd {
	fmt.Println("started job")

	// fileX, _ := CreateTightVNCFiles("45.55.153.222", "5901", "prime6996")

	runVNCcommand()

	return func() tea.Msg {
		return backgroundJobMsg{result: fmt.Sprintf("VNC file created - %s", "pepa")}
	}

}

func (m *MenuList) backgroundJobCreateBox() tea.Cmd {
	fmt.Println("started job")

	firewallID := createFirewall(doAPI)
	saveFirewall(firewallID)

	pepitaID := createBox(doAPI)
	addDropletToFirewall(doAPI, firewallID, pepitaID)
	saveIDs(strconv.Itoa(pepitaID))

	return func() tea.Msg {
		return backgroundJobMsg{result: fmt.Sprintf("Droplet created! ID:%d\n", pepitaID)}
	}

}

func (m *MenuList) backgroundJobDeleteBox() tea.Cmd {
	fmt.Println("started job")

	deleteALLboxes()

	return func() tea.Msg {
		return backgroundJobMsg{result: "Droplets Deleted!"}
	}

}

func saveFirewall(id string) error {
	execpath, _ := os.Executable()
	dir := filepath.Dir(execpath)
	filepath := filepath.Join(dir, "firewall.txt")

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 06044)

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	_, err = file.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func saveIDs(id string) error {
	execpath, _ := os.Executable()
	dir := filepath.Dir(execpath)
	filepath := filepath.Join(dir, "ids.txt")

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 06044)

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	_, err = file.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func getIDs() ([]string, error) {
	execpath, _ := os.Executable()
	dir := filepath.Dir(execpath)
	filepath := filepath.Join(dir, "ids.txt")

	file, err := os.Open(filepath)

	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	var ids []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ids, nil
}

func deleteALLboxes() {
	boxes, _ := getIDs()

	for _, str := range boxes {
		numID, _ := strconv.Atoi(str)
		deleteBox(doAPI, numID)
	}

}

func ShowMenuList(header string) {
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuMainColor))
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuMainColor))

	const defaultWidth = 90

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		list:       l,
		header:     header,
		state:      StateMainMenu,
		spinner:    s,
		spinnerMsg: "Action Performing",
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
