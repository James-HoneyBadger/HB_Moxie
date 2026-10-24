//go:build !windows

package main

import "syscall"

// setRestrictiveUmask keeps newly created database, log, and storage files
// private to the current user on multi-user systems.
func setRestrictiveUmask() {
	syscall.Umask(0077)
}
