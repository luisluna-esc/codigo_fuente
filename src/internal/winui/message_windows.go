//go:build windows

package winui

import (
	"syscall"
	"unsafe"
)

const (
	MBOK              = 0x00000000
	MBYesNo           = 0x00000004
	MBIconQuestion    = 0x00000020
	MBIconWarning     = 0x00000030
	MBIconInformation = 0x00000040
	MBDefButton2      = 0x00000100
	IDYes             = 6
	CSIDLPrograms     = 0x0002
	CSIDLDesktop      = 0x0010
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	msgBox          = user32.NewProc("MessageBoxW")
	shell32         = syscall.NewLazyDLL("shell32.dll")
	shGetFolderPath = shell32.NewProc("SHGetFolderPathW")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	createMutexW    = kernel32.NewProc("CreateMutexW")
	releaseMutex    = kernel32.NewProc("ReleaseMutex")
	closeHandle     = kernel32.NewProc("CloseHandle")
)

func Message(text, caption string, flags uintptr) int {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(caption)
	r, _, _ := msgBox.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), flags)
	return int(r)
}

func AskYesNo(text, caption string, warning bool) bool {
	flags := uintptr(MBYesNo | MBIconQuestion)
	if warning {
		flags = MBYesNo | MBIconWarning | MBDefButton2
	}
	return Message(text, caption, flags) == IDYes
}

func Folder(csidl int) string {
	buf := make([]uint16, 260)
	r, _, _ := shGetFolderPath.Call(0, uintptr(csidl), 0, 0, uintptr(unsafe.Pointer(&buf[0])))
	if r != 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

// AcquireSingleInstance keeps only one process associated with name alive at a time.
// The returned release function must be deferred by the caller when ok is true.
func AcquireSingleInstance(name string) (release func(), ok bool) {
	const errorAlreadyExists syscall.Errno = 183
	n, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return func() {}, true
	}
	h, _, callErr := createMutexW.Call(0, 1, uintptr(unsafe.Pointer(n)))
	if h == 0 {
		// Do not block installation if Windows unexpectedly refuses mutex creation.
		return func() {}, true
	}
	if errno, isErrno := callErr.(syscall.Errno); isErrno && errno == errorAlreadyExists {
		closeHandle.Call(h)
		return func() {}, false
	}
	return func() {
		releaseMutex.Call(h)
		closeHandle.Call(h)
	}, true
}
