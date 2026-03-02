package main

import (
	"os"
	"path/filepath"
	"vulta/initz"
	"vulta/pushz"
	"vulta/runz"
	"vulta/state"

	"strings"

	"github.com/spf13/cobra"
)

const Version = "0.3.1"

var nodeIp string
var quiet bool

func main() {
	var rootCmd = &cobra.Command{
		Use:   "vulta",
		Short: "Vulta is a high-performance deployment orchestrator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.Version = Version

	rootCmd.PersistentFlags().StringVarP(&nodeIp, "node", "n", "None", "Target Node IP")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet Mode")

	// Init
	var forceInit bool
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Vulta in the Current Directory",
		Run: func(cmd *cobra.Command, args []string) {
			initz.InitVulta(forceInit)
			// loadedConfig := initz.NewInventory().LoadVultaConf()
			// state.LoadDeploymentState().MakeStateFile(*loadedConfig)
		},
	}

	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Forcefully Initialize Vulta in the Current Directory")

	// push
	// var targetFiles []string
	var pushCmd = &cobra.Command{
		Use:   "push [file1 file2 ....]",
		Short: "Push Files to Remote Nodes",
		Run: func(cmd *cobra.Command, args []string) {
			loadedConfig := initz.NewInventory().LoadVultaConf()
			stateFile := filepath.Join(initz.NewInventory().ProjectRoot, ".vulta/state.json")
			_, err := os.Stat(stateFile)
			if err != nil {
				state.LoadDeploymentState().MakeStateFile(*loadedConfig)
			}
			pushz.PushFilesToRemote(loadedConfig, nodeIp, quiet, args)

		},
	}
	// pushCmd.Flags().StringSliceVarP(&targetFiles, "file", "f", []string{}, "Specific files to push")

	// run
	var runCmd = &cobra.Command{
		Use:   "run [command]",
		Short: "Run Commands on Remote Nodes",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := strings.Join(args, " ")
			loadedConfig := initz.NewInventory().LoadVultaConf()
			runz.RunCommand(loadedConfig, remoteCommand, nodeIp, quiet)
		},
	}

	// This allows the user to run 'vulta completion zsh' etc.
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				rootCmd.GenZshCompletion(os.Stdout)
				// ... add others as needed
			}
		},
	})

	rootCmd.AddCommand(initCmd, pushCmd, runCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// if action == "init" && len(args) > 2 {
// 	fmt.Println("Error: too many arguments for 'vulta init'.")
// 	fmt.Println("Usage: vulta init [--force|-f]")
// 	return
// }

// 	switch action {
// 	case "init":
// 		initz.InitVulta(forceInitPtr)
// 	case "push":
// 		loadedConfig := initz.NewInventory().LoadVultaConf()
// 		pushz.PushFilesToRemote(loadedConfig, nodeIp, quite)

// 	case "run":
// 		if len(args) < 2 {
// 			fmt.Println("Error: no remote command specified.")
// 			fmt.Println("Usage: vulta run <command>")
// 			return
// 		}
// 		remoteCommand := strings.Join(args[1:], " ")
// 		loadedConfig := initz.NewInventory().LoadVultaConf()
// 		runz.RunCommand(loadedConfig, remoteCommand, nodeIp, quite)

// 	default:
// 		fmt.Printf("Error: unknown command '%s'.\n", action)
// 		fmt.Println("Usage:")
// 		fmt.Println("  vulta init [--force|-f]")
// 		fmt.Println("  vulta push")
// 		fmt.Println("  vulta run <command>")
// 		fmt.Println("  vulta --version")
// 	}
// }
