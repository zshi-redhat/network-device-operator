package utils

import (
	"os"
)

const (
	ClusterTypeOpenshift = "openshift"
)

var ClusterType string

func init() {
	ClusterType = os.Getenv("CLUSTER_TYPE")
}
