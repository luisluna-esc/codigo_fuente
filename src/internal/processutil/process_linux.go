package processutil

import (
	"fmt"
	"os/exec"
	"syscall"
)

func Hidden(cmd *exec.Cmd)          { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
func Alive(pid int) bool            { return syscall.Kill(pid, 0) != syscall.ESRCH }
func Message(title, message string) { fmt.Println(title + ": " + message) }
