package cmd

import (
	"os"

	"github.com/protosio/app-store/db"
	"github.com/protosio/app-store/http"
	"github.com/protosio/app-store/util"

	"github.com/spf13/cobra"
)

var log = util.GetLogger()
var port = 8000

var rootCmd = &cobra.Command{
	Use:   "app-store",
	Short: "Protos app store for serving application installers",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the app store web server",
	Run: func(cmd *cobra.Command, args []string) {
		http.StartWebServer(port)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the database for the app store",
	Run: func(cmd *cobra.Command, args []string) {
		db.SetupDB()
	},
}

//Execute is the entry point to the command line menu
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	serveCmd.PersistentFlags().IntVarP(&port, "port", "p", 8000, "port to listen on")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
}
