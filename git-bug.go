//go:generate go run doc/gen_docs.go
//go:generate go run misc/gen_completion.go

package main

import (
	"github.com/daedaleanai/git-ticket/commands"
)

func main() {
	commands.Execute()
}
