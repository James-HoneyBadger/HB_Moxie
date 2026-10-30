package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *AdminApp) makeDashboard() fyne.CanvasObject {
	userCount := a.queryInt("SELECT COUNT(*) FROM user")
	activeUsers := a.queryInt("SELECT COUNT(*) FROM user WHERE disabled = 0 AND deleted = 0")
	disabledUsers := a.queryInt("SELECT COUNT(*) FROM user WHERE disabled = 1")
	deletedUsers := a.queryInt("SELECT COUNT(*) FROM user WHERE deleted = 1")
	accountCount := a.queryInt("SELECT COUNT(*) FROM account")
	txCount := a.queryInt("SELECT COUNT(*) FROM `transaction`")
	categoryCount := a.queryInt("SELECT COUNT(*) FROM transaction_category")
	tagCount := a.queryInt("SELECT COUNT(*) FROM transaction_tag")

	dbSize := "Unknown"
	if info, err := os.Stat(a.dbPath); err == nil {
		dbSize = formatBytes(info.Size())
	}

	title := widget.NewLabel("Honey Ledger System Administration")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	userCard := widget.NewCard("Users", "Account statistics",
		container.NewGridWithColumns(2,
			widget.NewLabel("Total"), boldLabel(fmt.Sprintf("%d", userCount)),
			widget.NewLabel("Active"), boldLabel(fmt.Sprintf("%d", activeUsers)),
			widget.NewLabel("Disabled"), boldLabel(fmt.Sprintf("%d", disabledUsers)),
			widget.NewLabel("Deleted"), boldLabel(fmt.Sprintf("%d", deletedUsers)),
		),
	)

	dataCard := widget.NewCard("Data", "Content statistics",
		container.NewGridWithColumns(2,
			widget.NewLabel("Accounts"), boldLabel(fmt.Sprintf("%d", accountCount)),
			widget.NewLabel("Transactions"), boldLabel(fmt.Sprintf("%d", txCount)),
			widget.NewLabel("Categories"), boldLabel(fmt.Sprintf("%d", categoryCount)),
			widget.NewLabel("Tags"), boldLabel(fmt.Sprintf("%d", tagCount)),
		),
	)

	systemCard := widget.NewCard("System Information", "",
		container.NewGridWithColumns(2,
			widget.NewLabel("Config File"),
			widget.NewLabel(a.configPath),
			widget.NewLabel("Database Path"),
			widget.NewLabel(a.dbPath),
			widget.NewLabel("Database Size"),
			widget.NewLabel(dbSize),
			widget.NewLabel("Database Type"),
			widget.NewLabel("SQLite3"),
		),
	)

	statsRow := container.NewGridWithColumns(2, userCard, dataCard)

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		statsRow,
		systemCard,
	)
}

func boldLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}
