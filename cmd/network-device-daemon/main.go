package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const (
	componentName = "network-device-daemon"
)

var (
	rootCmd = &cobra.Command{
		Use:   componentName,
		Short: "Run Network Device Daemon",
		Long:  "Runs the Network Device Daemon which handles network device configuration on the host",
	}
)

func init() {
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		glog.Exitf("Error executing mcd: %v", err)
	}
}
