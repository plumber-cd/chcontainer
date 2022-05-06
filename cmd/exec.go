package cmd

import (
	llog "log"

	"github.com/plumber-cd/chcontainer/k8s"
	"github.com/plumber-cd/chcontainer/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	execCmd = &cobra.Command{
		Use:                   "exec [chcontainer flags] container [container cmd] [-- [container args]]",
		Short:                 "Executes into a remote container",
		Long:                  "See https://github.com/plumber-cd/chcontainer/README.md for details",
		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Debug.Print("Start exec command execution")

			k8s.Run(args)
		},
	}
)

func init() {
	execCmd.Flags().BoolP("quiet", "q", false, `Enable quiet mode.
	By default chcontainer never prints to StdOut,
	reserving that channel exclusively to the container.
	But it does print messages to StdErr.
	Enabling quiet mode will redirect all messages to the info logger.
	If --log mode was not enabled - these messages will be discarded.`)
	if err := viper.BindPFlag("quiet", execCmd.Flags().Lookup("quiet")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().Bool("log", false, "Enables info logs to file")
	if err := viper.BindPFlag("log", execCmd.Flags().Lookup("log")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().Bool("debug", false, "Enables info and debug logs to file")
	if err := viper.BindPFlag("debug", execCmd.Flags().Lookup("debug")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().BoolP("stdin", "s", true, "Redirect StdIn to the container")
	if err := viper.BindPFlag("stdin", execCmd.Flags().Lookup("stdin")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().BoolP("tty", "t", true, "Enables TTY, disable if piping something to stdin")
	if err := viper.BindPFlag("tty", execCmd.Flags().Lookup("tty")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().StringP("pod", "p", "", "Specify pod name to exec to")
	if err := viper.BindPFlag("pod", execCmd.Flags().Lookup("pod")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().StringP("container", "c", "", "Specify container name to exec to")
	if err := viper.BindPFlag("container", execCmd.Flags().Lookup("container")); err != nil {
		llog.Panic(err)
	}

	execCmd.Flags().SetInterspersed(false)

	rootCmd.AddCommand(execCmd)
}
