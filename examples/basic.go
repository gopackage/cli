package main

import (
	"fmt"
	"github.com/badslug/cli"
)

func main() {
	program := cli.New()
	program.SetName("Device Tool")
	program.SetDescription("Device troubleshooting tool")
	program.SetVersion("0.1")

	program.Option("-v, --verbose", "display verbose information", "")

	program.Command("tcp <port>", "capture TCP packets on <port>").
		SetAction(func(program *cli.Program, command *cli.Command, unknownArgs []string) {
		fmt.Println("doing tcp")
	})

	program.Topic("path", "setting the path for reading")

	program.Parse()
}
