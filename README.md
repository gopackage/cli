# cli

  The complete solution for [Go](http://golang.org) command-line interfaces, inspired by Ruby's [commander](https://github.com/visionmedia/commander) and Node's [commander.js](https://github.com/visionmedia/commander.js).

## Installation

    $ go get github.com/gopackage/cli

## Default Command

A default command can be registered using the `*` command string.

## Option parsing

 Options with cli are defined with the `.Option()` method, also serving as documentation for the options. The example below parses args and options from `os.Args` and sets the `*.Value` for corresponding Options and Args on the program or command.

```go
package main

import (
  "fmt"
  "github.com/gopackage/cli"
)

func main() {
  program := cli.New()
  program.SetVersion("0.1")

  program.Command("tcp <port>", "capture TCP packets on <port>").
    SetAction(func(program *cli.Program, command *cli.Command, unknownArgs []string) {
      fmt.Printf("tcp on port %d", command.Args[0].IntValue())
  })

  program.Parse()
}
```

Short flags may be passed as a single arg, for example `-abc` is equivalent to `-a -b -c`. Long flags that start with `--no-` are automatically boolean options.

## Automated help

 The help information is auto-generated based on the information commander already knows about your program, so the following `help` info is for free:

```  
$ ./examples/capture help

   Usage: capture [options]

The commands are:

     version      output version number
     tcp <port>   capture TCP packets on <port>
     help [cmd]   display help for [cmd]

Use "capture help [command]" for more information about a command.

```

## Help topics

You can add custom help topics to document information relevant to the program (e.g. environmental variables) that aren't specific to a command.


```go

package main

import (
  "fmt"
  "github.com/gopackage/cli"
)

func main() {
  program := cli.New()
  program.SetVersion("0.1")

  program.Topic("path", "setting the path for reading").SetBody("Long form topic description of the path setting.")

  program.Parse()
}

```

Users will see the topic listed on the standard help output, and typing `capture help path` to see the `path` topic information.

## Custom help

 You can display arbitrary `help` information by registering
 your own help command. Your help command will override the built-in
 version. `cli` will exit once you are done so that the remainder of your program
 does not execute causing undesired behaviours, for example
 in the following executable "stuff" will not output when
 `help` is used.

```go

package main

import (
  "fmt"
  "github.com/gopackage/cli"
)

func main() {
  program := cli.New()
  program.SetVersion("0.1")

  program.Command("help [cmd]", "display help information for the program or a [cmd]").
    SetAction(func(program *cli.Program, command *cli.Command, unknownArgs []string) {
      fmt.Println("This is my custom help message")
  })

  program.Parse()
}

```

## .PrintHelp()

  Print help information without exiting.

## .Help()

  Print help information and exit immediately.

## Links

 - API documentation

# Developers

Command line processing is implemented in `cli.go` and terminal tools in
`terminal.go`.

## Testing

Tests use [Gomega](http://github.com/onsi/gomega) and the
[Ginkgo](http://github.com/onsi/ginkgo) assertion library. Tests can be
run using the standard Go testing functionality `go test .`

## License

(The MIT License)

Copyright (c) 2014-2017 Iain Shigeoka

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
