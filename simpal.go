/*
**
** simpal
** Provides a simple read-only, no ansi/escape/control code terminal with easy copy pasting
** to use with Fyne.
**
** Distributed under the COOL License.
**
** Copyright (c) 2024 IPv6.rs <https://ipv6.rs>
** All Rights Reserved
**
*/

package simpal

import (
  "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/canvas"
  "fyne.io/fyne/v2/container"
  "fyne.io/fyne/v2/driver/desktop"
  "fyne.io/fyne/v2/theme"
  "fyne.io/fyne/v2/widget"
  "fmt"
  "regexp"
  "runtime"
  "sync"
  "time"
)

type Terminal struct {
  app             fyne.App
  window          fyne.Window
  mu              *sync.Mutex
  lines           *[]string
  selectedItems   map[int]bool
  outputList      *widget.List
  timestamp       int64
  lastClickedItem int
  shiftHeld       bool
  ctrlHeld        bool
}


func NewTerminal(app fyne.App, window fyne.Window, command string) *container.Scroll {
  term := &Terminal{
    app:           app,
    window:        window,
    mu:            &sync.Mutex{},
    lines:         &[]string{},
    selectedItems: make(map[int]bool),
    timestamp:     time.Now().UnixMilli(),
  }

  term.outputList = setupOutputArea(term)
  scrollContainer := container.NewScroll(term.outputList)

  setupShortcuts(window, term)
  setupKeyHandlers(window, term)
  go executeCommand(term, command)

  return scrollContainer
}

func setupOutputArea(term *Terminal) *widget.List {
  list := widget.NewList(
    func() int {
      term.mu.Lock()
      defer term.mu.Unlock()
      return len(*term.lines)
    },
    func() fyne.CanvasObject {
      text := canvas.NewText("", theme.ForegroundColor())
      text.TextSize = 14
      text.TextStyle = fyne.TextStyle{Monospace: true}
      return text
    },
    func(id widget.ListItemID, object fyne.CanvasObject) {
      term.mu.Lock()
      defer term.mu.Unlock()
      text := object.(*canvas.Text)
      text.Text = (*term.lines)[id]
      if _, ok := term.selectedItems[id]; ok {
        text.Color = theme.PrimaryColor()
      } else {
        text.Color = theme.ForegroundColor()
      }
      object.Refresh()
    },
  )

  list.OnSelected = func(id widget.ListItemID) {
    fmt.Println("Selected item ID:", id)

    term.mu.Lock()
    if term.shiftHeld && term.lastClickedItem >= 0 {
      start, end := min(term.lastClickedItem, id), max(term.lastClickedItem, id)
      for i := start; i <= end; i++ {
        term.selectedItems[i] = true
      }
    } else if term.ctrlHeld {
      fmt.Println("CTRL held, toggling item")
      if _, selected := term.selectedItems[id]; selected {
        delete(term.selectedItems, id)
      } else {
        term.selectedItems[id] = true
        term.lastClickedItem = id
      }
    } else {
      term.selectedItems = make(map[int]bool)
      term.selectedItems[id] = true
      term.lastClickedItem = id
    }
    term.mu.Unlock()
    
    list.Refresh()
    fmt.Println("List refreshed")
    term.window.Canvas().Focus(nil)
  }

  return list
}

func (term *Terminal) addLine(text string, scroll bool) {
  term.mu.Lock()
  *term.lines = append(*term.lines, text)
  term.mu.Unlock()
  term.app.Driver().CanvasForObject(term.outputList).Content().Refresh()
  if scroll {
    term.outputList.ScrollToBottom()
  }
}

func (term *Terminal) copySelectedToClipboard() {
  term.mu.Lock()
  defer term.mu.Unlock()
  var text string
  for i, line := range *term.lines {
    if term.selectedItems[i] {
      text += line + "\n"
    }
  }
  if len(text) > 0 {
    term.window.Clipboard().SetContent(text)
  }
}

func stripANSICodes(text string) string {
  return regexp.MustCompile(`\x1b[\x40-\x5A\x5C\x5F\x60-\x7E]|\x1b\[[0-?]*[ -/]*[@-~]|(\x1b\][0-9]?;[^\x1b]*)(\x1b\\)?`).ReplaceAllString(text, "")
}

func setupShortcuts(window fyne.Window, term *Terminal) {
  shortcutCopy := &desktop.CustomShortcut{
    KeyName:  fyne.KeyC,
    Modifier: desktop.AltModifier,
  }
  window.Canvas().AddShortcut(shortcutCopy, func(shortcut fyne.Shortcut) {
    term.copySelectedToClipboard()
  })
}

func min(a, b int) int {
  if a < b {
    return a
  }
  return b
}

func max(a, b int) int {
  if a > b {
    return a
  }
  return b
}

func setupKeyHandlers(window fyne.Window, term *Terminal) {
  canvas := window.Canvas()
  if deskCanvas, ok := canvas.(desktop.Canvas); ok {
    deskCanvas.SetOnKeyDown(func(key *fyne.KeyEvent) {
      controlKeyLeft := desktop.KeyControlLeft
      controlKeyRight := desktop.KeyControlRight
      if runtime.GOOS == "darwin" {
        controlKeyLeft = desktop.KeySuperLeft
        controlKeyRight = desktop.KeySuperRight
      }
      switch key.Name {
        case desktop.KeyShiftLeft, desktop.KeyShiftRight:
          term.shiftHeld = true
          fmt.Println("Shift key pressed")
        case controlKeyLeft, controlKeyRight:
          term.ctrlHeld = true
          fmt.Println("Control/Command key pressed")
      }
    })
    deskCanvas.SetOnKeyUp(func(key *fyne.KeyEvent) {
      controlKeyLeft := desktop.KeyControlLeft
      controlKeyRight := desktop.KeyControlRight
      if runtime.GOOS == "darwin" {
        controlKeyLeft = desktop.KeySuperLeft
        controlKeyRight = desktop.KeySuperRight
      }

      switch key.Name {
        case desktop.KeyShiftLeft, desktop.KeyShiftRight:
          term.shiftHeld = false
          fmt.Println("Shift key released")
        case controlKeyLeft, controlKeyRight:
          term.ctrlHeld = false
          fmt.Println("Control/Command key released")
      }
    })
  } else {
    fmt.Println("Canvas does not support desktop keyboard events")
  }
}
