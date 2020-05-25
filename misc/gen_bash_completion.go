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
	dir := path.Join(cwd, "misc", "bash_completion", "git-ticket")

	fmt.Println("Generating Bash completion file ...")

	err := commands.RootCmd.GenBashCompletionFile(dir)
	if err != nil {
		log.Fatal(err)
	}
}
