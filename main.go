package main

import (
	"github.com/protosio/app-store/cmd"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

func main() {
	log.Info("Starting the Protos app store")
	cmd.Execute()
}
