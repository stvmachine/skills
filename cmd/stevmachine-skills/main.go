package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: stevmachine-skills <command> [args]
Commands:
  install [skill ...]   Install skills (default: all)
  list                  List installed skills
  doctor                Check installation health
  env set KEY VALUE     Store an encrypted variable
  env list              List variables (masked)
  env encrypt           Encrypt .env
  env decrypt           Decrypt .env
  env rotate            Rotate encryption keys
  env setup             Interactive TUI wizard`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall(os.Args[2:])
	case "list":
		cmdList()
	case "doctor":
		cmdDoctor()
	case "env":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "set":
			cmdEnvSet(os.Args[3:])
		case "list":
			cmdEnvList(os.Args[3:])
		case "encrypt":
			cmdEnvEncrypt(os.Args[3:])
		case "decrypt":
			cmdEnvDecrypt(os.Args[3:])
		case "rotate":
			cmdEnvRotate(os.Args[3:])
		case "setup":
			cmdEnvSetup()
		default:
			usage()
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}
