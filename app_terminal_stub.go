//go:build !unix

package main

import "errors"

func (a *App) TerminalStart(string) error {
	return errors.New("integrated terminal is not available on this platform")
}

func (a *App) TerminalStop() error {
	return nil
}

func (a *App) TerminalWrite(string) error {
	return errors.New("terminal not running")
}

func (a *App) TerminalResize(int, int) error {
	return nil
}
