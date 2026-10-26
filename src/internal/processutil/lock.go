package processutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// An old lock is removable only when its recorded process has ended. Unknown
// lock formats and processes we cannot inspect remain protected.
func ClearStaleLock(path string) error {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	line := strings.SplitN(string(b), "\n", 2)[0]
	pid, err := strconv.Atoi(strings.TrimPrefix(line, "pid="))
	if err != nil || pid <= 0 || Alive(pid) {
		return fmt.Errorf("el servidor está activo o su bloqueo requiere revisión; detén Música antes de actualizar")
	}
	return os.Remove(path)
}
