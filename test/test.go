/*
**
** simpal
** Example usage of simpal but also a usable simpal command
**
** Distributed under the COOL License.
**
** Copyright (c) 2024 IPv6.rs <https://ipv6.rs>
** All Rights Reserved
**
*/

package main

import (
  "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/app"
  "github.com/ipv6rslimited/simpal"
  "os"
  "strings"
)

func main() {
  a := app.New()

  w := a.NewWindow("simpal")
  w.Resize(fyne.NewSize(800, 400))

  command := "echo 'No command provided'"
  if len(os.Args) > 1 {
    command = strings.Join(os.Args[1:], " ")
  }

  w.SetContent(simpal.NewTerminal(a, w, command))

  w.ShowAndRun()
}
