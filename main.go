package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dma23/currencyconvertertui/currency"
	"github.com/dma23/currencyconvertertui/tui"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	resultStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")).
			PaddingLeft(2).
			MarginTop(1).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000")).
			MarginTop(1)
)

type model struct {
	spinner        spinner.Model
	loading        bool
	error          error
	formData       tui.UIData
	result         string
	currencyRates  map[string]float64
	showConversion bool
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner: s,
		loading: false,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if m.showConversion {
				m.showConversion = false
				return m, nil
			}
		}

	case tui.FormCompletedMsg:
		m.formData = msg.Data
		m.loading = true
		return m, tea.Batch(
			m.spinner.Tick,
			loadRates,
		)

	case ratesLoadedMsg:
		m.currencyRates = msg.rates
		m.error = msg.err
		m.loading = false

		if m.error == nil {
			// Perform conversion
			amount, _ := strconv.ParseFloat(m.formData.Amount, 64)
			rate := m.currencyRates[m.formData.Currency2] / m.currencyRates[m.formData.Currency1]
			convertedAmount := amount * rate

			m.result = fmt.Sprintf("%.2f %s = %.2f %s",
				amount,
				m.formData.Currency1,
				convertedAmount,
				m.formData.Currency2)
			m.showConversion = true
		}

		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if !m.showConversion {
		return "\n" + titleStyle.Render("Currency Converter") + "\n"
	}

	s := "\n" + titleStyle.Render("Currency Converter") + "\n\n"

	if m.loading {
		s += fmt.Sprintf("%s Loading currency rates...\n", m.spinner.View())
		return s
	}

	if m.error != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %v\n", m.error))
		s += "Using fallback rates.\n\n"
	}

	s += resultStyle.Render(m.result)
	s += "\n\nPress 'r' to convert again or 'q' to quit\n"

	return s
}

type ratesLoadedMsg struct {
	rates map[string]float64
	err   error
}

func loadRates() tea.Msg {
	rates, err := currency.GetRates()
	if err != nil {
		// Fallback to predefined rates
		rates = currency.Rates
	}
	return ratesLoadedMsg{rates: rates, err: err}
}

func main() {
	// Create the initial model
	m := initialModel()

	// Start with the form
	formData, err := tui.StartForm()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Update the model with form data
	m.formData = formData
	m.loading = true
	m.showConversion = true

	// Start the TUI with the updated model
	p := tea.NewProgram(m)

	// Start immediately with loading rates
	go func() {
		rates, err := currency.GetRates()
		if err != nil {
			// Fallback to predefined rates
			rates = currency.Rates
		}
		p.Send(ratesLoadedMsg{rates: rates, err: err})
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
