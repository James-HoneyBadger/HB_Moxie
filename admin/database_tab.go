package main

import (
	"database/sql"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *AdminApp) makeDatabaseTab() fyne.CanvasObject {
	tables := a.loadTables()
	pageSize := 50

	// Forward-declare so closures can reference it
	var tableListWidget *widget.List

	// --- Browse state ---
	var currentTable string
	currentPage := 1
	var dbColumns []string
	var dbRows [][]string

	// Browse data table
	dataTable := widget.NewTable(
		func() (int, int) {
			if len(dbRows) == 0 {
				return 0, 0
			}
			return len(dbRows), len(dbColumns)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("cell content placeholder")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row < len(dbRows) && id.Col < len(dbRows[id.Row]) {
				label.SetText(dbRows[id.Row][id.Col])
			}
		},
	)
	dataTable.ShowHeaderRow = true
	dataTable.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabel("")
		l.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
		return l
	}
	dataTable.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		if id.Row == -1 && id.Col >= 0 && id.Col < len(dbColumns) {
			label.SetText(dbColumns[id.Col])
		}
	}

	pageLabel := widget.NewLabel("No table selected")

	// Load data for selected table and page
	loadData := func() {
		if currentTable == "" {
			return
		}
		offset := (currentPage - 1) * pageSize
		quotedName := fmt.Sprintf("`%s`", currentTable)

		// Get columns via PRAGMA
		pragmaRows, err := a.db.Query("PRAGMA table_info(" + quotedName + ")")
		if err != nil {
			pageLabel.SetText("Error: " + err.Error())
			return
		}
		dbColumns = nil
		for pragmaRows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var dflt sql.NullString
			pragmaRows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
			dbColumns = append(dbColumns, name)
		}
		pragmaRows.Close()

		// Get data
		dataRows, err := a.db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", quotedName, pageSize, offset))
		if err != nil {
			pageLabel.SetText("Error: " + err.Error())
			return
		}
		dbRows = nil
		for dataRows.Next() {
			dbRows = append(dbRows, scanRowStrings(dataRows, len(dbColumns)))
		}
		dataRows.Close()

		// Update column widths based on header length
		for i, col := range dbColumns {
			w := float32(len(col)*10 + 40)
			if w < 80 {
				w = 80
			}
			if w > 300 {
				w = 300
			}
			dataTable.SetColumnWidth(i, w)
		}

		var total int64
		a.db.QueryRow("SELECT COUNT(*) FROM " + quotedName).Scan(&total)
		totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
		if totalPages == 0 {
			totalPages = 1
		}

		pageLabel.SetText(fmt.Sprintf("Table: %s  |  Page %d / %d  |  %d total rows",
			currentTable, currentPage, totalPages, total))
		dataTable.Refresh()
	}

	prevBtn := widget.NewButton("← Previous", func() {
		if currentPage > 1 {
			currentPage--
			loadData()
		}
	})
	nextBtn := widget.NewButton("Next →", func() {
		currentPage++
		loadData()
	})
	firstBtn := widget.NewButton("First", func() {
		currentPage = 1
		loadData()
	})

	pagination := container.NewHBox(firstBtn, prevBtn, layout.NewSpacer(), pageLabel, layout.NewSpacer(), nextBtn)
	browseView := container.NewBorder(nil, pagination, nil, nil, dataTable)

	// --- Query (Read-Only) state ---
	var qrColumns []string
	var qrRows [][]string

	queryResultTable := widget.NewTable(
		func() (int, int) {
			if len(qrRows) == 0 {
				return 0, 0
			}
			return len(qrRows), len(qrColumns)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("cell placeholder")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row < len(qrRows) && id.Col < len(qrRows[id.Row]) {
				label.SetText(qrRows[id.Row][id.Col])
			}
		},
	)
	queryResultTable.ShowHeaderRow = true
	queryResultTable.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabel("")
		l.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
		return l
	}
	queryResultTable.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		if id.Row == -1 && id.Col >= 0 && id.Col < len(qrColumns) {
			label.SetText(qrColumns[id.Col])
		}
	}

	queryInput := widget.NewMultiLineEntry()
	queryInput.SetPlaceHolder("SELECT * FROM user LIMIT 10;")
	queryInput.SetMinRowsVisible(3)
	queryInput.TextStyle = fyne.TextStyle{Monospace: true}

	queryStatus := widget.NewLabel("Enter a SELECT or PRAGMA query (max 1,000 rows)")

	executeQueryBtn := widget.NewButton("Execute Query", func() {
		query := strings.TrimSpace(queryInput.Text)
		if query == "" {
			return
		}
		upper := strings.ToUpper(query)
		if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "PRAGMA") {
			queryStatus.SetText("Error: Only SELECT and PRAGMA queries allowed here. Use the Execute tab for writes.")
			return
		}

		rows, err := a.db.Query(query)
		if err != nil {
			queryStatus.SetText("Error: " + err.Error())
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		qrColumns = cols
		qrRows = nil

		count := 0
		for rows.Next() && count < 1000 {
			qrRows = append(qrRows, scanRowStrings(rows, len(cols)))
			count++
		}

		// Auto-size columns
		for i, col := range qrColumns {
			w := float32(len(col)*10 + 40)
			if w < 80 {
				w = 80
			}
			if w > 300 {
				w = 300
			}
			queryResultTable.SetColumnWidth(i, w)
		}

		queryStatus.SetText(fmt.Sprintf("%d rows returned", count))
		queryResultTable.Refresh()
	})
	executeQueryBtn.Importance = widget.HighImportance

	clearQueryBtn := widget.NewButton("Clear", func() {
		queryInput.SetText("")
		qrColumns = nil
		qrRows = nil
		queryResultTable.Refresh()
		queryStatus.SetText("Enter a SELECT or PRAGMA query (max 1,000 rows)")
	})

	queryToolbar := container.NewHBox(executeQueryBtn, clearQueryBtn, layout.NewSpacer(), queryStatus)
	queryView := container.NewBorder(
		container.NewVBox(queryInput, queryToolbar),
		nil, nil, nil,
		queryResultTable,
	)

	// --- Execute (Write) state ---
	updateInput := widget.NewMultiLineEntry()
	updateInput.SetPlaceHolder("UPDATE user SET nickname = 'Admin' WHERE uid = 1;")
	updateInput.SetMinRowsVisible(3)
	updateInput.TextStyle = fyne.TextStyle{Monospace: true}

	updateStatus := widget.NewLabel("")

	warningLabel := widget.NewLabel("⚠ Warning: Statements executed here directly modify the database. There is no undo.")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}

	executeUpdateBtn := widget.NewButton("Execute Statement", func() {
		query := strings.TrimSpace(updateInput.Text)
		if query == "" {
			return
		}

		dialog.ShowConfirm("Execute SQL Statement",
			"Are you sure you want to execute:\n\n"+query+"\n\nThis cannot be undone.",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				result, err := a.db.Exec(query)
				if err != nil {
					updateStatus.SetText("Error: " + err.Error())
					return
				}

				affected, _ := result.RowsAffected()
				updateStatus.SetText(fmt.Sprintf("Success: %d rows affected", affected))

				// Refresh browse data and table list
				if currentTable != "" {
					loadData()
				}
				tables = a.loadTables()
				if tableListWidget != nil {
					tableListWidget.Refresh()
				}
			}, a.window)
	})
	executeUpdateBtn.Importance = widget.DangerImportance

	clearUpdateBtn := widget.NewButton("Clear", func() {
		updateInput.SetText("")
		updateStatus.SetText("")
	})

	updateToolbar := container.NewHBox(executeUpdateBtn, clearUpdateBtn, layout.NewSpacer(), updateStatus)
	updateView := container.NewVBox(
		warningLabel,
		updateInput,
		updateToolbar,
	)

	// --- Export ---
	exportBtn := widget.NewButton("Export Database as SQL Dump", func() {
		d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()
			a.exportDatabase(writer)
			dialog.ShowInformation("Export Complete",
				"Database exported successfully to:\n"+writer.URI().Path(),
				a.window)
		}, a.window)
		d.SetFileName(fmt.Sprintf("ratel_ledger_export_%s.sql", ""))
		d.Show()
	})
	exportBtn.Importance = widget.HighImportance

	exportView := container.NewVBox(
		widget.NewLabel("Export the entire database as a SQL dump file."),
		widget.NewLabel("The export includes all table schemas and data as INSERT statements."),
		widget.NewSeparator(),
		exportBtn,
	)

	// --- Right-side tabs ---
	rightTabs := container.NewAppTabs(
		container.NewTabItem("Browse", browseView),
		container.NewTabItem("Query (Read-Only)", queryView),
		container.NewTabItem("Execute (Write)", updateView),
		container.NewTabItem("Export", exportView),
	)

	// --- Table list (left panel) ---
	tableListWidget = widget.NewList(
		func() int { return len(tables) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("table_name_placeholder"),
				layout.NewSpacer(),
				widget.NewLabel("0"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(tables[id].Name)
			box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("%d", tables[id].RowCount))
		},
	)

	tableListWidget.OnSelected = func(id widget.ListItemID) {
		currentTable = tables[id].Name
		currentPage = 1
		loadData()
	}

	refreshTablesBtn := widget.NewButton("Refresh Tables", func() {
		tables = a.loadTables()
		tableListWidget.Refresh()
	})

	leftPanel := container.NewBorder(
		widget.NewLabel("Tables"),
		refreshTablesBtn,
		nil, nil,
		tableListWidget,
	)

	// Split layout
	split := container.NewHSplit(leftPanel, rightTabs)
	split.Offset = 0.18

	return split
}
