// +build !windows

/*
**
** simpal
** Executes in darwin and linux
** This would not be possible without Fyne Terminal as a guide.
**
** Distributed under the COOL License.
**
** Copyright (c) 2024 IPv6.rs <https://ipv6.rs>
** All Rights Reserved
**
*/

package simpal

import (
  "time"
  "bufio"
  "fmt"
  "os"
  "os/exec"
  "github.com/creack/pty"
)

func executeCommand(term *Terminal, command string) {
  time.Sleep(500 * time.Millisecond)

  cmd := exec.Command("bash", "-c", command)
  cmd.Env = append(os.Environ(), "TERM=dumb", "PATH=" + os.Getenv("PATH"))

  ptmx, err := pty.Start(cmd)
  if err != nil {
    term.addLine(fmt.Sprintf("Error starting pty: %s", err.Error()), false)
    return
  }
  defer ptmx.Close()

  updatePTYSize(ptmx)

  scanner := bufio.NewScanner(ptmx)
  for scanner.Scan() {
    line := stripANSICodes(scanner.Text())
    term.addLine(line, true)
  }
}

func updatePTYSize(ptmx *os.File) {
  if ptmx == nil {
    return
  }
  _ = pty.Setsize(ptmx, &pty.Winsize{
    Rows: 48,  
    Cols: 1000,
  })
}


