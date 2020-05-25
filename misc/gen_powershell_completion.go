// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/daedaleanai/git-ticket/commands"
)

func main() {
	cwd, _ := os.Getwd()
	filepath := path.Join(cwd, "misc", "powershell_completion", "git-ticket")

	fmt.Println("Generating PowerShell completion file ...")

	err := commands.RootCmd.GenPowerShellCompletionFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
}
