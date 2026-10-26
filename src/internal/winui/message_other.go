//go:build !windows

package winui

import "fmt"

const (
	MBOK              = 0
	MBYesNo           = 4
	MBIconQuestion    = 0x20
	MBIconWarning     = 0x30
	MBIconInformation = 0x40
	MBDefButton2      = 0x100
	CSIDLPrograms     = 0x0002
	CSIDLDesktop      = 0x0010
)

func Message(text, caption string, flags uintptr) int {
	fmt.Printf("%s: %s\n", caption, text)
	return 1
}
func AskYesNo(text, caption string, warning bool) bool {
	fmt.Printf("%s: %s\n", caption, text)
	return true
}
func Folder(csidl int) string { return "" }
func AcquireSingleInstance(name string) (release func(), ok bool) {
	return func() {}, true
}
