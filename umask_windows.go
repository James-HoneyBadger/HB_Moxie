//go:build windows

package main

// setRestrictiveUmask is a no-op on Windows, which uses ACLs instead of umask.
func setRestrictiveUmask() {}
