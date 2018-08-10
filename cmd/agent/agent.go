package main

import (
	"fmt"
	"os"

	"github.com/moooofly/hunter-agent/cli"
	"github.com/moooofly/hunter-agent/daemon/config"
	"github.com/moooofly/hunter-agent/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func runDaemon(opts *daemonOptions) error {
	daemonCli := NewDaemonCli()
	return daemonCli.start(opts)
}

func newDaemonCommand() *cobra.Command {
	opts := newDaemonOptions(config.New())

	cmd := &cobra.Command{
		Use:           "hunter-agent [OPTIONS]",
		Short:         "This is an agent for hunter system as a proxy.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.flags = cmd.Flags()
			return runDaemon(opts)
		},
		DisableFlagsInUseLine: true,
		Version:               fmt.Sprintf("%s, build %s", version.Version, version.GitCommit),
	}
	cli.SetupRootCommand(cmd)

	flags := cmd.Flags()
	flags.BoolP("version", "v", false, "Print version information and quit")
	flags.StringVar(&opts.configFile, "config-file", defaultDaemonConfigFile, "Daemon configuration file")
	opts.InstallFlags(flags)
	installConfigFlags(opts.daemonConfig, flags)

	return cmd
}

func main() {
	logrus.SetOutput(os.Stderr)
	cmd := newDaemonCommand()
	cmd.SetOutput(os.Stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
