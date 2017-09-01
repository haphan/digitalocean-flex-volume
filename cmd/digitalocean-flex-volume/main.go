package main

import (
	"fmt"
	"os"

	"github.com/StackPointCloud/digitalocean-flex-volume/cmd/digitalocean-flex-volume/config"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/digitalocean/cloud"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/digitalocean/plugin"
	"github.com/StackPointCloud/digitalocean-flex-volume/pkg/flex"
	"github.com/golang/glog"
)

// func init() {
// 	flag.Set("logtostderr", "true")
// }
//
// func debugFile(msg string) {
// 	file := "/var/log/digitalocean.log"
// 	var f *os.File
//
// 	if _, err := os.Stat(file); os.IsNotExist(err) {
// 		f, err = os.Create(file)
// 		if err != nil {
// 			panic(err)
// 		}
// 	} else {
// 		f, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	defer f.Close()
//
// 	if _, err := f.WriteString(msg + "\n"); err != nil {
// 		panic(err)
// 	}
// }

func main() {

	// Create the digital ocean manager
	token, err := config.GetDigitalOceanToken()
	if err != nil {
		glog.Errorf("Error retrieving Digital Ocean token: %v", err.Error())
		os.Exit(1)
	}

	glog.Info("Creating Digital Ocean client")
	do, err := cloud.NewDigitalOceanManager(token)
	if err != nil {
		glog.Errorf("Error creating Digital Ocean client: %v", err.Error())
		os.Exit(1)
	}

	// create Digital Ocean flex volume instance
	p := plugin.NewDigitalOceanVolumePlugin(do)

	// create flex Executor
	manager := flex.NewManager(p, os.Stdout)

	// read arguments
	args := os.Args
	if len(args) < 2 {
		manager.WriteError(fmt.Errorf("flex command argument was not found"))
		os.Exit(1)
	}

	// create flex command based on flags
	fc, err := flex.NewFlexCommand(args)
	if err != nil {
		manager.WriteError(err)
		os.Exit(1)
	}

	// execute flex command
	ds, err := manager.ExecuteCommand(fc)
	if err != nil {
		manager.WriteError(err)
		os.Exit(1)
	}

	// write result to output
	err = manager.WriteDriverStatus(ds)
	if err != nil {
		manager.WriteError(err)
		os.Exit(1)
	}
}
