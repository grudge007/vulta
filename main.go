package main

import (
	"flag"
	"fmt"
	"vulta/initz"
	"vulta/pushz"
	"vulta/runz"

	"strings"
)

const Version = "0.1.1"

func main() {
	var action string
	var forceInitPtr bool
	var nodePtr string
	var versionPtr bool

	flag.StringVar(&nodePtr, "node", "None", "Node IP")
	flag.StringVar(&nodePtr, "n", "None", "Node IP")

	flag.BoolVar(&versionPtr, "version", false, "Vesrion")
	flag.BoolVar(&versionPtr, "v", false, "Vesrion")

	flag.BoolVar(&forceInitPtr, "force", false, "Force Init")
	flag.BoolVar(&forceInitPtr, "f", false, "Force Init")

	flag.Parse()

	if versionPtr {
		fmt.Printf("vulta version %s\n", Version)
		return
	}

	args := flag.Args()

	if len(args) > 0 {
		action = args[0]
	} else {
		fmt.Println("Usage:")
		fmt.Println("  vulta init [--force|-f]")
		fmt.Println("  vulta push")
		fmt.Println("  vulta -node <node ip> push")
		fmt.Println("  vulta run <command>")
		fmt.Println("  vulta -node <node ip> run <command>")
		fmt.Println("  vulta --version")
		return
	}

	nodeIp := nodePtr

	if action == "init" && len(args) > 2 {
		fmt.Println("Error: too many arguments for 'vulta init'.")
		fmt.Println("Usage: vulta init [--force|-f]")
		return
	}

	switch action {
	case "init":
		initz.InitVulta(forceInitPtr)
	case "push":
		loadedConfig := initz.NewInventory().LoadVultaConf()
		pushz.PushFilesToRemote(loadedConfig, nodeIp)

	case "run":
		if len(args) < 2 {
			fmt.Println("Error: no remote command specified.")
			fmt.Println("Usage: vulta run <command>")
			return
		}
		remoteCommand := strings.Join(args[1:], " ")
		loadedConfig := initz.NewInventory().LoadVultaConf()
		runz.RunCommand(loadedConfig, remoteCommand, nodeIp)

	default:
		fmt.Printf("Error: unknown command '%s'.\n", action)
		fmt.Println("Usage:")
		fmt.Println("  vulta init [--force|-f]")
		fmt.Println("  vulta push")
		fmt.Println("  vulta run <command>")
		fmt.Println("  vulta --version")
	}
}
