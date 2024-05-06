// +build windows

/* 
**
** simpal
** Executes in windows
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
  "bufio"
  "fmt"
  "os"
  "syscall"
  "github.com/ActiveState/termtest/conpty"
  "time"
)

func executeCommand(term *Terminal, command string) {
  time.Sleep(500 * time.Millisecond)

  cpty, err := conpty.New(80, 24)
  if err != nil {
     term.addLine(fmt.Sprintf("Error starting pty: %s", err.Error()), false)
     return
  }
  defer cpty.Close()

  pid, _, err := cpty.Spawn("C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe", []string{"-NoLogo", "-NoExit", "-Command", command}, &syscall.ProcAttr{
    Env: append(os.Environ(), "TERM=dumb"),
  })
  if err != nil {
    term.addLine(fmt.Sprintf("Error spawning command: %s", err.Error()), false)
    return
  }

  updatePTYSize(cpty)

  process, err := os.FindProcess(pid)
  if err != nil {
    term.addLine(fmt.Sprintf("Error finding process: %s", err.Error()), false)
    return
  }
  defer process.Release()

  outPipe := cpty.OutPipe()

  scanner := bufio.NewScanner(outPipe)
  for scanner.Scan() {
    line := stripANSICodes(scanner.Text())
    term.addLine(line, true)
  }

  if _, err := process.Wait(); err != nil {
    term.addLine(fmt.Sprintf("Error waiting for process: %s", err.Error()), false)
  }
}

func updatePTYSize(pty *conpty.ConPty) {
    if pty == nil {
        return
    }
    _ = pty.Resize(1000, 48)
}

