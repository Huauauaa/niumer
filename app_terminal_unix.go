//go:build unix

package main

import (
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func terminalShellPath() string {
	if sh := strings.TrimSpace(os.Getenv("SHELL")); sh != "" {
		return sh
	}
	if fi, err := os.Stat("/bin/zsh"); err == nil && !fi.IsDir() {
		return "/bin/zsh"
	}
	return "/bin/bash"
}

func terminalResolveCwd(a *App, cwd string) string {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		cwd = a.GetBlogWorkDir()
	}
	if st, err := os.Stat(cwd); err != nil || !st.IsDir() {
		if h, err := os.UserHomeDir(); err == nil && h != "" {
			return h
		}
		return "."
	}
	return cwd
}

// terminalStopLocked closes the PTY master and kills the child. The session goroutine
// calls Wait and clears a.terminalCmd when this shell is still the active one.
// Caller must hold a.muTerminal.
func (a *App) terminalStopLocked() {
	if a.terminalFile != nil {
		_ = a.terminalFile.Close()
		a.terminalFile = nil
	}
	if a.terminalCmd != nil && a.terminalCmd.Process != nil {
		_ = a.terminalCmd.Process.Kill()
	}
}

func (a *App) forwardTerminalOutput(fd *os.File) {
	buf := make([]byte, 32768)
	for {
		n, err := fd.Read(buf)
		if n > 0 && a.ctx != nil {
			runtime.EventsEmit(a.ctx, "terminal:output", base64.StdEncoding.EncodeToString(buf[:n]))
		}
		if err != nil {
			if a.ctx != nil && !errors.Is(err, io.EOF) {
				runtime.EventsEmit(a.ctx, "terminal:error", err.Error())
			}
			return
		}
	}
}

func (a *App) runTerminalSession(cmd *exec.Cmd, fd *os.File) {
	defer func() {
		if cmd != nil {
			_ = cmd.Wait()
		}
		a.muTerminal.Lock()
		if a.terminalCmd == cmd {
			a.terminalCmd = nil
			a.terminalFile = nil
		}
		a.muTerminal.Unlock()
	}()
	a.forwardTerminalOutput(fd)
}

// TerminalStart spawns an interactive shell in a PTY. cwd empty uses blog work dir then home.
func (a *App) TerminalStart(cwd string) error {
	cwd = terminalResolveCwd(a, cwd)

	a.muTerminal.Lock()
	defer a.muTerminal.Unlock()

	a.terminalStopLocked()

	cmd := exec.Command(terminalShellPath())
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	a.terminalCmd = cmd
	a.terminalFile = f
	go a.runTerminalSession(cmd, f)

	log.Printf("niumer: terminal started in %s (%s)", cwd, terminalShellPath())
	return nil
}

// TerminalStop tears down the shell and PTY.
func (a *App) TerminalStop() error {
	a.muTerminal.Lock()
	defer a.muTerminal.Unlock()
	a.terminalStopLocked()
	return nil
}

// TerminalWrite sends raw bytes (UTF-8 from xterm) to the shell.
func (a *App) TerminalWrite(data string) error {
	a.muTerminal.Lock()
	f := a.terminalFile
	a.muTerminal.Unlock()
	if f == nil {
		return errors.New("terminal not running")
	}
	_, err := f.Write([]byte(data))
	return err
}

// TerminalResize updates PTY dimensions (cols × rows).
func (a *App) TerminalResize(cols, rows int) error {
	if cols < 1 || rows < 1 {
		return nil
	}
	a.muTerminal.Lock()
	f := a.terminalFile
	a.muTerminal.Unlock()
	if f == nil {
		return nil
	}
	return pty.Setsize(f, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}
