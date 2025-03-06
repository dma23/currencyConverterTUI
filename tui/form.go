package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/dma23/currencyconvertertui/currency"
)

type UIData struct {
	Amount    string
	Currency1 string
	Currency2 string
}

func StartForm() (UIData, error) {

	var data UIData = UIData{
		Amount:    "1.00",
		Currency1: "CAD",
		Currency2: "USD",
	}

	options := currency.CurrencyTypes

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Amount").
				Value(&data.Amount).
				Validate(func(s string) error {
					_, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return fmt.Errorf("please enter a valid number")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("From Currency").
				Options(
					buildCurrencyOptions(options)...,
				).
				Value(&data.Currency1),
			huh.NewSelect[string]().
				Title("To Currency").
				Options(
					buildCurrencyOptions(options)...,
				).
				Value(&data.Currency2),
		),
	).WithTheme(huh.ThemeCharm())

	// Run the form
	err := form.Run()
	return data, err
}

func buildCurrencyOptions(codes map[string]string) []huh.Option[string] {
	var options []huh.Option[string]
	for _, code := range codes {
		displayText := fmt.Sprintf("%s", code)
		options = append(options, huh.NewOption(displayText, code))
	}
	return options
}
