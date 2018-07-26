package main

import (
	"github.com/spf13/cobra"

	"github.com/anhnguyentb/cws-publish/cmds/cws-upload"
	"github.com/anhnguyentb/cws-publish/cmds/store-config"
	"github.com/anhnguyentb/cws-publish/tools"
)

const name = "cws-publish"

func main() {
	log := tools.InitLogging()
	rootCmd := &cobra.Command{
		Use:   name,
		Short: "Includes tools to build & publish Chrome Web Store",
	}
	rootCmd.AddCommand(store_config.InitCommand(log))
	rootCmd.AddCommand(cws_upload.InitCommand(log))
	tools.PreExecuteConfiguration(rootCmd, name, log)
	tools.Execute(rootCmd, log)
}
