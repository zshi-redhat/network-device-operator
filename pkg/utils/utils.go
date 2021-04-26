package utils

import (
	"os"
	"syscall"
)

const (
	ClusterTypeOpenshift = "openshift"
)

var ClusterType string

func init() {
	ClusterType = os.Getenv("CLUSTER_TYPE")
}

func Chroot(path string) (func() error, error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, err
	}

	if err := syscall.Chroot(path); err != nil {
		root.Close()
		return nil, err
	}

	return func() error {
		defer root.Close()
		if err := root.Chdir(); err != nil {
			return err
		}
		return syscall.Chroot(".")
	}, nil
}
