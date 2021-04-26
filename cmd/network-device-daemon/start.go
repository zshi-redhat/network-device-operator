package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/zshi-redhat/network-device-operator/pkg/daemon"
	"github.com/zshi-redhat/network-device-operator/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts Network Device Daemon",
		Long:  "",
		Run:   runStartCmd,
	}

	startOpts struct {
		kubeconfig string
		nodeName   string
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringVar(&startOpts.kubeconfig, "kubeconfig", "", "Kubeconfig file to access a remote cluster (testing only)")
	startCmd.PersistentFlags().StringVar(&startOpts.nodeName, "node-name", "", "kubernetes node name daemon is managing.")
}

func runStartCmd(cmd *cobra.Command, args []string) {
	flag.Set("logtostderr", "true")
	flag.Parse()

	if startOpts.nodeName == "" {
		name, ok := os.LookupEnv("NODE_NAME")
		if !ok || name == "" {
			glog.Fatalf("Node name is required")
		}
		startOpts.nodeName = name
	}

	// This channel is used to ensure all spawned goroutines exit when we exit.
	stopCh := make(chan struct{})
	defer close(stopCh)

	var config *rest.Config
	var err error

	if os.Getenv("CLUSTER_TYPE") == utils.ClusterTypeOpenshift {
		kubeconfig, err := clientcmd.LoadFromFile("/host/etc/kubernetes/kubeconfig")
		if err != nil {
			glog.Errorf("Failed to load kubelet kubeconfig: %v", err)
		}
		clusterName := kubeconfig.Contexts[kubeconfig.CurrentContext].Cluster
		apiURL := kubeconfig.Clusters[clusterName].Server

		url, err := url.Parse(apiURL)
		if err != nil {
			glog.Errorf("Failed to parse api url from kubelet kubeconfig: %v", err)
		}

		// The kubernetes in-cluster functions don't let you override the apiserver
		// directly; gotta "pass" it via environment vars.
		glog.V(0).Infof("overriding kubernetes api to %s", apiURL)
		os.Setenv("KUBERNETES_SERVICE_HOST", url.Hostname())
		os.Setenv("KUBERNETES_SERVICE_PORT", url.Port())
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		panic(err.Error())
	}

	kubeclient := kubernetes.NewForConfigOrDie(config)

	glog.V(0).Info("starting NetworkDeviceDaemon")
	err = daemon.New(
		startOpts.nodeName,
		stopCh,
		kubeclient,
	).Run()
	if err != nil {
		glog.Errorf("Failed to run daemon: %v", err)
	}
	glog.V(0).Info("shutting down SriovNetworkConfigDaemon")
}
