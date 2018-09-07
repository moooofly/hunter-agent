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
		Short:         "This is an agent running as a proxy of Hunter system.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.flags = cmd.Flags()
			return runDaemon(opts)
		},
		DisableFlagsInUseLine: true,
		Version: fmt.Sprintf("| % -20s | % -20s |\n| % -20s | % -20s |\n| % -20s | % -20s |\n",
			"version", version.Version,
			"git commit", version.GitCommit,
			"build time", version.BuildTime),
	}
	cli.SetupRootCommand(cmd)

	flags := cmd.Flags()
	flags.BoolP("version", "v", false, "Print version information and quit")
	opts.InstallCommonOptionsFlags(flags)
	installDaemonConfigFlags(opts.daemonConfig, flags)

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
