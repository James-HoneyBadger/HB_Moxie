package main

import (
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
)

func (a *AdminApp) makeConfigTab() fyne.CanvasObject {
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, a.configPath)
	if err != nil {
		return widget.NewLabel("Error loading config: " + err.Error())
	}

	// Collect section names (skip DEFAULT)
	var sectionNames []string
	for _, s := range cfg.Sections() {
		if s.Name() != "DEFAULT" {
			sectionNames = append(sectionNames, s.Name())
		}
	}

	// Pre-create Entry widgets for every key in every section
	// entryMap[section][key] = *widget.Entry
	type configItem struct {
		key   string
		entry *widget.Entry
	}
	allItems := map[string][]configItem{}

	for _, sectionName := range sectionNames {
		section := cfg.Section(sectionName)
		for _, key := range section.Keys() {
			var entry *widget.Entry
			if isPasswordKey(key.Name()) {
				entry = widget.NewPasswordEntry()
			} else {
				entry = widget.NewEntry()
			}
			entry.SetText(key.Value())
			allItems[sectionName] = append(allItems[sectionName], configItem{
				key:   key.Name(),
				entry: entry,
			})
		}
	}

	// Right-side form area (initially empty)
	formBox := container.NewVBox(
		widget.NewLabel("Select a configuration section from the list"),
	)
	formScroll := container.NewVScroll(formBox)

	// Section list
	sectionList := widget.NewList(
		func() int { return len(sectionNames) },
		func() fyne.CanvasObject { return widget.NewLabel("section_name_placeholder") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText("[" + sectionNames[id] + "]")
		},
	)

	sectionList.OnSelected = func(id widget.ListItemID) {
		sectionName := sectionNames[id]
		items := allItems[sectionName]

		var formItems []*widget.FormItem
		for _, item := range items {
			formItems = append(formItems, widget.NewFormItem(item.key, item.entry))
		}

		header := widget.NewLabel("[" + sectionName + "]")
		header.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

		form := widget.NewForm(formItems...)

		formBox.Objects = []fyne.CanvasObject{header, form}
		formBox.Refresh()
	}

	// Status label
	statusLabel := widget.NewLabel("")

	// Save button
	saveBtn := widget.NewButton("Save Configuration", func() {
		dialog.ShowConfirm("Save Configuration",
			"Save changes to honey-ledger.ini?\nA backup will be created before saving.",
			func(ok bool) {
				if !ok {
					return
				}

				// Create backup
				backupPath := a.configPath + ".bak." + time.Now().Format("20060102-150405")
				if data, err := os.ReadFile(a.configPath); err == nil {
					if err := os.WriteFile(backupPath, data, 0600); err != nil {
						dialog.ShowError(fmt.Errorf("failed to create backup: %w", err), a.window)
						return
					}
				}

				// Load fresh config to preserve comments and formatting
				newCfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, a.configPath)
				if err != nil {
					dialog.ShowError(err, a.window)
					return
				}

				// Apply all entry values
				for sectionName, items := range allItems {
					for _, item := range items {
						newCfg.Section(sectionName).Key(item.key).SetValue(item.entry.Text)
					}
				}

				if err := newCfg.SaveTo(a.configPath); err != nil {
					dialog.ShowError(err, a.window)
					return
				}

				statusLabel.SetText("Saved! Backup: " + backupPath)
				dialog.ShowInformation("Configuration Saved",
					"Configuration saved successfully.\nBackup created at:\n"+backupPath+
						"\n\nRestart Honey Ledger for changes to take effect.",
					a.window)
			}, a.window)
	})
	saveBtn.Importance = widget.HighImportance

	// Reset button
	resetBtn := widget.NewButton("Reset to File", func() {
		dialog.ShowConfirm("Reset", "Discard all changes and reload from file?", func(ok bool) {
			if !ok {
				return
			}
			freshCfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, a.configPath)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			for sectionName, items := range allItems {
				section := freshCfg.Section(sectionName)
				for i, item := range items {
					val := section.Key(item.key).Value()
					items[i].entry.SetText(val)
				}
			}
			statusLabel.SetText("Reset to file values")
		}, a.window)
	})

	toolbar := container.NewHBox(saveBtn, resetBtn, layout.NewSpacer(), statusLabel)

	// Layout: toolbar on top, split panel in center
	split := container.NewHSplit(sectionList, formScroll)
	split.Offset = 0.18

	return container.NewBorder(toolbar, nil, nil, nil, split)
}
