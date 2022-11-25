package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Note struct {
	file    string
	changed bool
}

func (n *Note) Reset() {
	n.file = ""
	n.changed = false
}

func (n *Note) Save(data string, parent fyne.Window) bool {
	if n.file == "" {
		//Not created yet
		log.Panicln("Note file is empty - ask")
	}

	log.Printf("Writing to %v: %v", n.file, data)
	file, err := os.Create(n.file)
	defer file.Close()

	if err != nil {
		log.Printf("Could not create file %v", n.file)
		dialog.ShowError(err, parent)
		return false
	}

	bytedata := []byte(data)
	_, werr := file.Write(bytedata)
	if werr != nil {
		log.Printf("Error writing to %v: %v", n.file, werr)
		dialog.ShowError(werr, parent)
		return false
	}

	n.changed = false
	return true
}

func make_new_note(n *Note, w fyne.Window) {
	n.Reset()
	w.SetTitle("New Note")
}

func save_note(n *Note, data string, w fyne.Window) {
	if n.file == "" {
		//Not created yet
		log.Println("Note file is empty - ask")
		filedialog := dialog.NewFileSave(func(io fyne.URIWriteCloser, dialog_err error) {
			log.Println("NewFileSave callback")
			if dialog_err == nil {
				n.file = io.URI().Path()
				log.Printf("Will save to %v", n.file)
				n.Save(data, w)
			} else {
				log.Println("File name selection cancelled")
			}
		}, w)
		filedialog.SetFileName("NewNote.txt")
		filedialog.Show()
	} else {
		n.Save(data, w)
	}
}

func open_note(n *Note, settext func(data string), w fyne.Window) {
	filedialog := dialog.NewFileOpen(func(io fyne.URIReadCloser, dialog_err error) {
		if dialog_err == nil {
			content, err := ioutil.ReadFile(io.URI().Path())
			if err == nil {
				n.file = io.URI().Path()
				settext(string(content))
				n.changed = false
			}
		}
	}, w)
	filedialog.Show()
}

func main() {
	note := Note{
		file:    "",
		changed: false,
	}
	//a := app.New()
	a := app.NewWithID("com.example.noteeditor")
	w := a.NewWindow("Note Editor")

	footer := widget.NewLabel("---")
	body := widget.NewMultiLineEntry()

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			log.Println("Open note")
			if note.changed == true {
				dialog.NewConfirm("Discard Changes?", "The current note has not been saved. Discard changes?", func(choice bool) {
					if choice == true {
						open_note(&note, func(data string) {
							body.SetText(data)
						}, w)
					}
				}, w).Show()
			} else {
				open_note(&note, func(data string) {
					body.SetText(data)
				}, w)
			}
		}),
		widget.NewToolbarAction(theme.FileTextIcon(), func() {
			log.Println("New note")
			if note.changed == true {
				dialog.NewConfirm("Discard Changes?", "The current note has not been saved. Discard changes?", func(choice bool) {
					if choice == true {
						body.SetText("")
						make_new_note(&note, w)
					}
				}, w).Show()
			} else {
				body.SetText("")
				make_new_note(&note, w)
			}
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			log.Println("Save")
			if note.changed == true {
				save_note(&note, body.Text, w)
			} else {
				log.Println("Already saved")
			}
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			log.Println("Clear contents")
			if len(body.Text) > 0 && note.changed == true {
				dialog.NewConfirm("Discard Changed?", "The current note has not been saved. Are you sure?", func(choice bool) {
					if choice == true {
						body.SetText("")
						note.changed = true
					}
				}, w).Show()
			} else {
				body.SetText("")
				note.changed = true
			}
		}),
	)

	body.OnChanged = func(input string) {
		footer.SetText(fmt.Sprintf("%v symbols", len(input)))
		note.changed = true
	}

	blayout := container.New(layout.NewBorderLayout(toolbar, footer, nil, nil), toolbar, footer, body)

	if _, ok := a.(desktop.App); ok {
		w.SetMainMenu(fyne.NewMainMenu(
			fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { w.Close() })),
			fyne.NewMenu("Help", fyne.NewMenuItem("About", func() {
				dialog.ShowInformation("About", "Demo note editor v 0.0.1", w)
			})),
		))
	}
	w.SetContent(blayout)
	w.Resize(fyne.NewSize(640, 480))
	w.ShowAndRun()
}
