package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *AdminApp) makeUsersTab() fyne.CanvasObject {
	users := a.loadUsers()
	selectedRow := -1

	headers := []string{"UID", "Username", "Email", "Nickname", "Currency", "Language", "Status", "Email Verified", "Created", "Last Login"}
	colWidths := []float32{180, 120, 200, 120, 80, 80, 80, 100, 150, 150}

	table := widget.NewTable(
		func() (int, int) { return len(users), len(headers) },
		func() fyne.CanvasObject {
			l := widget.NewLabel("template text here")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row >= len(users) {
				return
			}
			u := users[id.Row]
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", u.Uid))
			case 1:
				label.SetText(u.Username)
				label.TextStyle = fyne.TextStyle{Bold: true}
			case 2:
				label.SetText(u.Email)
			case 3:
				label.SetText(u.Nickname)
			case 4:
				label.SetText(u.DefaultCurrency)
			case 5:
				label.SetText(u.Language)
			case 6:
				if u.Deleted {
					label.SetText("Deleted")
				} else if u.Disabled {
					label.SetText("Disabled")
				} else {
					label.SetText("Active")
				}
			case 7:
				if u.EmailVerified {
					label.SetText("Yes")
				} else {
					label.SetText("No")
				}
			case 8:
				label.SetText(formatUnixTime(u.CreatedUnixTime))
			case 9:
				label.SetText(formatUnixTime(u.LastLoginUnixTime))
			}
		},
	)

	// Set column widths
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}

	// Header row
	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabel("")
		l.TextStyle = fyne.TextStyle{Bold: true}
		return l
	}
	table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		if id.Row == -1 && id.Col >= 0 && id.Col < len(headers) {
			label.SetText(headers[id.Col])
		}
	}

	table.OnSelected = func(id widget.TableCellID) {
		selectedRow = id.Row
	}

	// Status label
	statusLabel := widget.NewLabel(fmt.Sprintf("%d users loaded", len(users)))

	// Refresh helper
	refresh := func() {
		users = a.loadUsers()
		selectedRow = -1
		table.Refresh()
		statusLabel.SetText(fmt.Sprintf("%d users loaded", len(users)))
	}

	getSelectedUID := func() (int64, string, bool) {
		if selectedRow < 0 || selectedRow >= len(users) {
			dialog.ShowInformation("No Selection", "Please select a user first.", a.window)
			return 0, "", false
		}
		return users[selectedRow].Uid, users[selectedRow].Username, true
	}

	// Action buttons
	refreshBtn := widget.NewButton("Refresh", func() { refresh() })

	enableBtn := widget.NewButton("Enable", func() {
		uid, name, ok := getSelectedUID()
		if !ok {
			return
		}
		_, err := a.db.Exec("UPDATE user SET disabled = 0, updated_unix_time = ? WHERE uid = ?",
			time.Now().Unix(), uid)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to enable user: %w", err), a.window)
			return
		}
		statusLabel.SetText(fmt.Sprintf("Enabled user: %s", name))
		refresh()
	})

	disableBtn := widget.NewButton("Disable", func() {
		uid, name, ok := getSelectedUID()
		if !ok {
			return
		}
		dialog.ShowConfirm("Disable User",
			fmt.Sprintf("Disable user '%s'?", name),
			func(confirmed bool) {
				if !confirmed {
					return
				}
				_, err := a.db.Exec("UPDATE user SET disabled = 1, updated_unix_time = ? WHERE uid = ?",
					time.Now().Unix(), uid)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to disable user: %w", err), a.window)
					return
				}
				statusLabel.SetText(fmt.Sprintf("Disabled user: %s", name))
				refresh()
			}, a.window)
	})

	deleteBtn := widget.NewButton("Delete", func() {
		uid, name, ok := getSelectedUID()
		if !ok {
			return
		}
		dialog.ShowConfirm("Delete User",
			fmt.Sprintf("Soft-delete user '%s'? The user can be restored later.", name),
			func(confirmed bool) {
				if !confirmed {
					return
				}
				now := time.Now().Unix()
				_, err := a.db.Exec("UPDATE user SET deleted = 1, deleted_unix_time = ?, updated_unix_time = ? WHERE uid = ?",
					now, now, uid)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to delete user: %w", err), a.window)
					return
				}
				statusLabel.SetText(fmt.Sprintf("Deleted user: %s", name))
				refresh()
			}, a.window)
	})
	deleteBtn.Importance = widget.DangerImportance

	restoreBtn := widget.NewButton("Restore", func() {
		uid, name, ok := getSelectedUID()
		if !ok {
			return
		}
		_, err := a.db.Exec("UPDATE user SET deleted = 0, deleted_unix_time = 0, updated_unix_time = ? WHERE uid = ?",
			time.Now().Unix(), uid)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to restore user: %w", err), a.window)
			return
		}
		statusLabel.SetText(fmt.Sprintf("Restored user: %s", name))
		refresh()
	})

	verifyBtn := widget.NewButton("Verify Email", func() {
		uid, name, ok := getSelectedUID()
		if !ok {
			return
		}
		_, err := a.db.Exec("UPDATE user SET email_verified = 1, updated_unix_time = ? WHERE uid = ?",
			time.Now().Unix(), uid)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to verify email: %w", err), a.window)
			return
		}
		statusLabel.SetText(fmt.Sprintf("Email verified for: %s", name))
		refresh()
	})

	toolbar := container.NewHBox(
		refreshBtn,
		widget.NewSeparator(),
		enableBtn, disableBtn,
		widget.NewSeparator(),
		deleteBtn, restoreBtn,
		widget.NewSeparator(),
		verifyBtn,
		layout.NewSpacer(),
		statusLabel,
	)

	return container.NewBorder(toolbar, nil, nil, nil, table)
}
