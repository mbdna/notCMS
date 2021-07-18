package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
)

var items []fyne.CanvasObject
var data = []string{}
var dataCurrent = []string{}

type shortcutEntry struct {
	widget.Entry
}

func newShortcutEntry() *shortcutEntry {
	entry := &shortcutEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (m *shortcutEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		m.Entry.TypedShortcut(s)
		return
	}

	if s.ShortcutName() == "CustomDesktop:Control+S" {
		SaveFile()

	}
}

func SaveFile() {
	fmt.Println("Save")
	b, err := ioutil.ReadFile("html/index.html") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	str := string(b)
	for i := 0; i < len(data); i++ {
		if data[i] != dataCurrent[i] {
			fmt.Println("Replace " + data[i] + " to " + dataCurrent[i])
			str = strings.Replace(str, data[i], dataCurrent[i], -1)
		}

	}
	fi, err := os.Create("html/index.html")
	fi.WriteString(str)
	fi.Sync()
}

func indexOf(objects []fyne.CanvasObject, object fyne.CanvasObject) int {

	for i := 0; i < len(objects); i++ {
		if objects[i] == object {
			return i
		}
	}
	return -1
}

func makeButtonList() []fyne.CanvasObject {

	for i := 0; i < len(data); i++ {
		index := i // capture
		entry := newShortcutEntry()
		entry.MultiLine = true
		entry.SetText(data[index])
		entry.OnChanged = func(input string) {
			index = indexOf(items, entry)
			dataCurrent[index] = input
			fmt.Println("old: " + data[index])
			fmt.Println("new: " + dataCurrent[index])
		}
		items = append(items, entry)
	}

	return items
}

func makeScrollTab() fyne.CanvasObject {
	vlist := makeButtonList()
	vert := container.NewVScroll(container.NewVBox(vlist...))
	saveButton := widget.NewButton(("Save"), func() {
		SaveFile()
	})
	return container.NewAdaptiveGrid(1,
		container.NewBorder(saveButton, nil, nil, nil, vert))
}

func startServer() {
	http.Handle("/", http.FileServer(http.Dir("./html")))
	http.ListenAndServe(":3000", nil)
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {
	f := app.New()
	w := f.NewWindow("")

	b, err := ioutil.ReadFile("html/index.html") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	str := string(b) // convert content to a 'string'
	paragraphStart, _ := regexp.Compile("<p>")
	paragraphEnd, _ := regexp.Compile("</p>")
	paragraphStartIndex := paragraphStart.FindAllStringIndex(str, -1)
	paragraphEndIndex := paragraphEnd.FindAllStringIndex(str, -1)

	for i := 0; i < len(paragraphStartIndex); i++ {
		fmt.Println(str[paragraphStartIndex[i][1]:paragraphEndIndex[i][0]])
		data = append(data, str[paragraphStartIndex[i][1]:paragraphEndIndex[i][0]])
		dataCurrent = append(dataCurrent, str[paragraphStartIndex[i][1]:paragraphEndIndex[i][0]])
	}

	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		fmt.Println("Save")
	})

	go startServer()
	open("http://localhost:3000/")

	w.SetContent(makeScrollTab())
	w.ShowAndRun()

}
