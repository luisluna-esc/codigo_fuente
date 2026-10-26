package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"sgl.local/sistema-gestion-liderazgo/internal/installstate"
	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
	"sgl.local/sistema-gestion-liderazgo/internal/winui"
)

const title = "Desinstalar Ministerio de la Música"

func main() {
	workerRoot := flag.String("worker-root", "", "uso interno del desinstalador")
	parent := flag.Int("parent-pid", 0, "uso interno del desinstalador")
	flag.Parse()
	exe, err := os.Executable()
	if err == nil && *workerRoot != "" {
		// The worker is outside the installation, allowing Windows to remove the
		// original executable once its parent has released it.
		for deadline := time.Now().Add(15 * time.Second); *parent > 0 && processutil.Alive(*parent); {
			if time.Now().After(deadline) {
				err = fmt.Errorf("cierre la otra ventana del desinstalador y reintente")
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if err == nil {
			err = uninstall(*workerRoot)
		}
		if runtime.GOOS == "windows" {
			deleteWorker(exe)
		}
	} else if err == nil {
		var root string
		root, err = installstate.CanonicalRoot(filepath.Dir(exe))
		if err == nil {
			_, err = installstate.Load(root)
		}
		if err == nil && !winui.AskYesNo("¿Desea desinstalar Ministerio de la Música Online?\n\nSe detendrá el servidor y se retirarán los archivos del programa.\nSus registros, certificados, respaldos, fuentes personalizadas y configuraciones se conservarán en:\n"+root+"\n\nPodrá recuperarlos reinstalando en esa misma carpeta.", title, true) {
			return
		}
		if err == nil {
			err = startWorker(exe, root)
		}
	}
	if err != nil {
		processutil.Message(title, "No se completó la desinstalación:\n"+err.Error()+"\n\nSus datos y configuraciones se conservaron. Puede volver a intentarlo.")
		os.Exit(1)
	}
}

func startWorker(exe, root string) error {
	temp, err := os.MkdirTemp("", "Musica-desinstalar-")
	if err != nil {
		return err
	}
	path := filepath.Join(temp, installstate.UninstallerName)
	in, err := os.Open(exe)
	if err != nil {
		_ = os.RemoveAll(temp)
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		_ = os.RemoveAll(temp)
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.RemoveAll(temp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.RemoveAll(temp)
		return closeErr
	}
	cmd := exec.Command(path, "-worker-root", root, "-parent-pid", fmt.Sprint(os.Getpid()))
	cmd.Dir = temp
	processutil.Hidden(cmd)
	if err = cmd.Start(); err != nil {
		_ = os.RemoveAll(temp)
		return err
	}
	return cmd.Process.Release()
}

func uninstall(destination string) error {
	root, err := installstate.CanonicalRoot(destination)
	if err != nil {
		return err
	}
	m, err := installstate.Load(root)
	if err != nil {
		return err
	}
	unlock, ok := winui.AcquireSingleInstance("Local\\MusicaUninstall-" + filepath.Base(root))
	if !ok {
		return fmt.Errorf("ya hay otra desinstalación en curso")
	}
	defer unlock()
	if err = installstate.RemoveStartupTask(root); err != nil {
		return err
	}
	if err = installstate.StopServer(root); err != nil {
		return err
	}
	preserved, err := m.RemoveProgram()
	if err != nil {
		return fmt.Errorf("cierre las ventanas de Música y reintente: %w", err)
	}
	if err = installstate.Unregister(root); err != nil {
		return err
	}
	if err = m.Finish(); err != nil {
		return err
	}
	message := "El programa se desinstaló.\n\nSus datos, certificados, respaldos y configuraciones se conservaron en:\n" + root + "\n\nPara volver a usarlos, reinstale en esa misma carpeta."
	if len(preserved) > 0 {
		message += "\n\nTambién se conservaron los archivos del programa que habían sido modificados."
	}
	processutil.Message(title, message)
	return nil
}
