package utils

import "syscall"

func StopApp() error {
	return syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
