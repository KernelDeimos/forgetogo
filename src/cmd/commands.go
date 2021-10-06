package cmd

import (
	"fmt"
	"os"

	"github.com/dube-dev/forgetogo/src/application"
	"github.com/dube-dev/forgetogo/src/managers/config"
	"github.com/dube-dev/forgetogo/src/managers/instance"
	"github.com/dube-dev/forgetogo/src/managers/server"
	"github.com/dube-dev/forgetogo/src/managers/tray"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "forgetogo",
	Run: func(cmd *cobra.Command, args []string) {
		app := application.NewApplication()
		app.AddManager("config", config.NewConfigManager())
		app.AddManager("tray", tray.NewTrayManager(app.QuitCh))
		app.AddManager("server", server.NewServerManager())
		app.AddManager("instance", instance.NewInstanceManager())
		app.Start()
		<-app.QuitCh
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
