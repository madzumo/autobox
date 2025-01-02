package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

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
	lipHeaderStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
	lipManifestStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	lipSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("205"))
	lipTitleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205"))

	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("111"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	// menuSMTPcolor       = "184"
	textPromptColor     = "141" //"100" //nice: 141
	textInputColor      = "193" //"40" //nice: 193
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textResultJob       = "141" //PINK"205"
	textJobOutcomeFront = "216"
	// txtJobOutcomeBack   = "205"

	menuTOP = []string{
		"Toggle Provider",
		"Enter API Token or Keys",
		"Set URL",
		"Change # of Boxes to deploy",
		"DEPLOY Boxes",
		"CREATE PS1 Scripts",
		"RUN Post URL Action",
		"VERIFY Boxes (TightVNC)",
		"DELETE All Boxes",
		"Save Settings",
	}
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
	app                 *applicationMain
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
					err := m.app.runPS1files()
					if err != nil {
						fmt.Printf("Error executing ps1 scripts:\n%s", err)
					}
					time.Sleep(10 * time.Second)

					m.app.updateHeader()
					m.header = m.app.header
					return m, nil
				case menuTOP[1]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., dop_v1_a0xx"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[2]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[2]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., https://www.whatever.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 200
					m.textInput.Width = 200
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[3]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuTOP[3]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 5"
					m.textInput.Focus()
					m.textInput.CharLimit = 10
					m.textInput.Width = 10
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuTOP[4]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobCreateBox())
				case menuTOP[5]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobPS1scripts())
				case menuTOP[6]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobRunPostURL())
					// m.prevMenuState = m.state
					// m.prevState = m.state
					// m.state = StateTextInput
					// m.inputPrompt = menuTOP[6]
					// m.textInput = textinput.New()
					// m.textInput.Placeholder = "e.g., Type All or box #"
					// m.textInput.Focus()
					// m.textInput.CharLimit = 10
					// m.textInput.Width = 10
					// m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					// m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					// return m, nil
				case menuTOP[7]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobVerifyVNC())
				case menuTOP[8]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundJobDeleteBox())
				case menuTOP[9]:
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.backgroundSaveSettings())
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
			inputValue := m.textInput.Value() // User pressed enter, save the input

			switch m.inputPrompt {
			case menuTOP[1]:
				m.app.settings.DoAPI = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved API: %s", inputValue)
				m.app.updateHeader()
				m.header = m.app.header
			case menuTOP[2]:
				m.app.settings.URL = inputValue
				m.backgroundJobResult = fmt.Sprintf("Saved URL: %s", inputValue)
				m.app.updateHeader()
				m.header = m.app.header
			case menuTOP[3]:
				boxes, err := strconv.Atoi(inputValue)
				if err != nil {
					m.backgroundJobResult = "Data inputed is not a valid Number"
				} else {
					m.app.settings.NumberBoxes = boxes
					m.backgroundJobResult = fmt.Sprintf("Number of Boxes = %s", inputValue)
					m.app.updateHeader()
					m.header = m.app.header
				}
				// case menuTOP[6]:
				// 	runBox := inputValue
				// 	if runBox == "all"{
				// 		return m, tea.Batch(m.spinner.Tick, m.backgroundJobRunPostURL())
				// 	}else{

				// 	}

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

func (m *MenuList) backgroundSaveSettings() tea.Cmd {
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	m.spinnerMsg = "Saving Settings"
	// m.spinner.Tick()
	time.Sleep(1 * time.Second)
	saveSettings(m.app.settings)

	return func() tea.Msg {
		return backgroundJobMsg{result: "Settings Saved"}
	}
}

func (m *MenuList) backgroundJobCreateBox() tea.Cmd {
	fmt.Println("started job")
	resultX := fmt.Sprintf("%d - Droplets created!", m.app.settings.NumberBoxes)

	err1 := m.app.createFirewall(m.app.settings.DoAPI)
	if err1 != nil {
		resultX = fmt.Sprintf("Error creating firewall:\n%s", err1)
	}

	for i := 1; i <= m.app.settings.NumberBoxes; i++ {
		err := m.app.createBox(m.app.settings.DoAPI)
		if err != nil {
			resultX = fmt.Sprintf("error creating box:\n%s", err)
		}
		time.Sleep(1 * time.Second)
	}

	return func() tea.Msg {
		return backgroundJobMsg{result: resultX}
	}

}

func (m *MenuList) backgroundJobRunPostURL() tea.Cmd {
	result := "Finished Post URL Execution"
	err := m.app.runPS1files()
	if err != nil {
		result = fmt.Sprintf("Error executing ps1 scripts:\n%s", err)
	}
	time.Sleep(10 * time.Second)

	return func() tea.Msg {
		return backgroundJobMsg{result: result}
	}
}

func (m *MenuList) backgroundJobPS1scripts() tea.Cmd {
	result := "Created PS1 files under Boxes folder"
	ips, err := m.app.compileIPaddresses()
	if err != nil {
		result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
	} else {
		startCount, err := countNumberofFiles("./boxes")
		if err != nil {
			fmt.Printf("Error getting number of Files\n%s", err)
			startCount = 0
		} else {
			fmt.Printf("Count of files: %d", startCount)
		}
		for id, ip := range ips {
			err := m.app.createPostSCRIPT(ip, (startCount + (id + 1)))
			if err != nil {
				result = fmt.Sprintf("Error creating post script\n%s", err)
			}
		}
	}

	return func() tea.Msg {
		return backgroundJobMsg{result: result}
	}
}

func (m *MenuList) backgroundJobDeleteBox() tea.Cmd {
	fmt.Println("started job")
	resultX := "Droplets Deleted!"

	err := m.app.deleteBox(m.app.settings.DoAPI)
	if err != nil {
		resultX = fmt.Sprintf("Error deleting droplets\n%s", err)
	}

	err = m.app.deleteFirewall(m.app.settings.DoAPI)
	if err != nil {
		resultX = fmt.Sprintf("Error deleting firewall\n%s", err)
	}

	err = os.RemoveAll("./boxes")
	if err != nil {
		resultX = fmt.Sprintf("Failed to delete boxes folder\n%s", err)
	}
	return func() tea.Msg {
		return backgroundJobMsg{result: resultX}
	}
}

func (m *MenuList) backgroundJobVerifyVNC() tea.Cmd {
	fmt.Println("started job")
	result := "Verified mofo!"

	ips, err := m.app.compileIPaddresses()
	if err != nil {
		result = fmt.Sprintf("Error compiling IP addresses:\n%s", err)
	} else {
		for _, ip := range ips {
			err := m.app.runVNC(ip)
			if err != nil {
				result = fmt.Sprintf("Error running TightVNC\n%s", err)
			}
		}
	}

	return func() tea.Msg {
		return backgroundJobMsg{result: result}
	}
}
func ShowMenu(app *applicationMain) {

	const defaultWidth = 90

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, defaultWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = lipTitleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		list:       l,
		header:     app.header,
		state:      StateMainMenu,
		spinner:    s,
		spinnerMsg: "Action Performing",
		app:        app,
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)

	//show Menu
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
