//go:build !windows

package installstate

func Register(root, version, icon string) error { return nil }
func RemoveStartupTask(root string) error       { return nil }
func Unregister(root string) error              { return nil }
