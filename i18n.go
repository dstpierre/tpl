package tpl

import (
	"fmt"
	"time"
)

// ToDate formats a date to a short date without time based on locale.
func ToDate(locale string, date time.Time) string {
	layout := "01-02-2006"

	switch locale {
	case "fr-CA", "en-CA":
		layout = "02-01-2006"
	}

	return date.Format(layout)
}

// ToCurrency formats an amounts based on locale with the proper currency sign.
func ToCurrency(locale string, amount float64) string {
	format := "$%.2f"

	switch locale {
	case "en-CA", "fr-CA":
		format = "%.2f $"
	}

	return fmt.Sprintf(format, amount)
}
