package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/golang/glog"
	ndv1alpha1 "github.com/zshi-redhat/network-device-operator/api/v1alpha1"
)

const (
	ClusterTypeOpenshift = "openshift"
	DeviceConfigFile     = "/host/etc/netdevice.conf"
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

func WriteDeviceConfFile(ndp *ndv1alpha1.NetDevicePool) (bool, error) {
	var err error
	_, err = os.Stat(DeviceConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			glog.V(2).Infof("WriteDeviceConfFile(): file not existed, create it")
			_, err = os.Create(DeviceConfigFile)
			if err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}
	newContent, err := json.Marshal(ndp.Spec.Device)
	if err != nil {
		return false, err
	}
	oldContent, err := ioutil.ReadFile(DeviceConfigFile)
	if err != nil {
		return false, err
	}
	if string(newContent) == string(oldContent) {
		glog.V(2).Info("WriteDeviceConfFile(): no update")
		return false, nil
	}
	glog.V(2).Infof("WriteDeviceConfFile(): write %s to /etc/netdevice.conf", newContent)
	err = ioutil.WriteFile(DeviceConfigFile, []byte(newContent), 0666)
	if err != nil {
		return false, err
	}
	return true, nil
}
