package main

import (
	_ "time/tzdata"

	"github.com/Sagarmikeylevi/Pulse-Sever/cmd/api/commands"
)

func main() {
	commands.Execute()
}
