// 2013 Iain Shigeoka - BSD license (see LICENSE)
package cli

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (s *CommandSuite) TestOptions(c *C) {
	option := NewOption(nil, "-v", "display version information")

	c.Check(option.Short, Equals, "-v")
	c.Check(option.Long, Equals, "")
	c.Check(option.Required, Equals, false)
	c.Check(option.Optional, Equals, false)
	c.Check(option.Bool, Equals, false)
	c.Check(option.Description, Equals, "display version information")

	option = NewOption(nil, "-v, --version", "display version information")

	c.Check(option.Short, Equals, "-v")
	c.Check(option.Long, Equals, "--version")
	c.Check(option.Required, Equals, false)
	c.Check(option.Optional, Equals, false)
	c.Check(option.Bool, Equals, false)
	c.Check(option.Description, Equals, "display version information")

	option = NewOption(nil, "-c, --config <path>", "set configuration file")

	c.Check(option.Short, Equals, "-c")
	c.Check(option.Long, Equals, "--config")
	c.Check(option.Required, Equals, true)
	c.Check(option.Optional, Equals, false)
	c.Check(option.Bool, Equals, false)
	c.Check(option.Description, Equals, "set configuration file")

	option = NewOption(nil, "-c, --config [path]", "set configuration file")

	c.Check(option.Short, Equals, "-c")
	c.Check(option.Long, Equals, "--config")
	c.Check(option.Required, Equals, false)
	c.Check(option.Optional, Equals, true)
	c.Check(option.Bool, Equals, false)
	c.Check(option.Description, Equals, "set configuration file")

	option = NewOption(nil, "-T, --no-tests", "ignore tests")

	c.Check(option.Short, Equals, "-T")
	c.Check(option.Long, Equals, "--no-tests")
	c.Check(option.Required, Equals, false)
	c.Check(option.Optional, Equals, false)
	c.Check(option.Bool, Equals, true)
	c.Check(option.Description, Equals, "ignore tests")
}

func (s *CommandSuite) TestCommands(c *C) {
	command := NewCommand(nil, "foo", "bar")

	c.Check(command.Command, Equals, "foo")
	c.Check(command.Description, Equals, "bar")

	command = NewCommand(nil, "foo <bar>", "a foo bar command")

	c.Check(command.Command, Equals, "foo")
	c.Check(command.Args, HasLen, 1)
	c.Check(command.Args[0].Name, Equals, "bar")
	c.Check(command.Args[0].Required, Equals, true)
	c.Check(command.Description, Equals, "a foo bar command")

	command = NewCommand(nil, "foo [bar]", "a foo bar command")

	c.Check(command.Command, Equals, "foo")
	c.Check(command.Args, HasLen, 1)
	c.Check(command.Args[0].Name, Equals, "bar")
	c.Check(command.Args[0].Required, Equals, false)
	c.Check(command.Description, Equals, "a foo bar command")
}

func (s *CommandSuite) TestNormalizeArgs(c *C) {
	program := New()

	normalized := program.normalize([]string{"help"})

	c.Assert(normalized, HasLen, 1)
	c.Check(normalized[0], Equals, "help")

	normalized = program.normalize([]string{"-v"})

	c.Assert(normalized, HasLen, 1)
	c.Check(normalized[0], Equals, "-v")

	normalized = program.normalize([]string{"--version"})

	c.Assert(normalized, HasLen, 1)
	c.Check(normalized[0], Equals, "--version")

	normalized = program.normalize([]string{"-abc"})

	c.Assert(normalized, HasLen, 3)
	c.Check(normalized[0], Equals, "-a")
	c.Check(normalized[1], Equals, "-b")
	c.Check(normalized[2], Equals, "-c")

	normalized = program.normalize([]string{"--port", "8080"})

	c.Assert(normalized, HasLen, 2)
	c.Check(normalized[0], Equals, "--port")
	c.Check(normalized[1], Equals, "8080")

	normalized = program.normalize([]string{"--port=8080"})

	c.Assert(normalized, HasLen, 2)
	c.Check(normalized[0], Equals, "--port")
	c.Check(normalized[1], Equals, "8080")
}

func (s *CommandSuite) TestOptionFor(c *C) {
	program := New()

	option := program.optionFor("-v")
	c.Check(option, IsNil)

	program.Option("-v, --version", "display option")

	option = program.optionFor("-v")
	c.Check(option.Name, Equals, "version")

	option = program.optionFor("--version")
	c.Check(option.Name, Equals, "version")

	option = program.optionFor("-f")
	c.Check(option, IsNil)
	option = program.optionFor("--foo")
	c.Check(option, IsNil)
}

func (s *CommandSuite) TestParseOptions(c *C) {
	program := New()
	args, unknown := program.parseOptions([]string{"help"})

	c.Assert(args, HasLen, 1)
	c.Check(args[0], Equals, "help")
	c.Assert(unknown, HasLen, 0)

	args, unknown = program.parseOptions([]string{"--foo"})

	c.Assert(args, HasLen, 0)
	c.Assert(unknown, HasLen, 1)
	c.Check(unknown[0], Equals, "--foo")

	program.Option("-f, --foo", "add a foo", "")

	args, unknown = program.parseOptions([]string{"--foo"})

	c.Assert(args, HasLen, 0)
	c.Assert(unknown, HasLen, 0)
}

func (s *CommandSuite) TestParseArgs(c *C) {
	program := New()
	program.SetDescription("Device troubleshooting tool")

	program.Option("-v, --verbose", "display verbose information")

	program.Command("tcp <port>", "capture TCP packets on <port>").Option("-h, --host", "host address to bind to")

	program.Topic("path", "setting the path for reading")

	command := program.parseArgs([]string{"help"}, []string{})

	c.Assert(command, IsNil)

	command = program.parseArgs([]string{"tcp", "8080"}, []string{})

	c.Assert(command, NotNil)
	c.Check(command.Command, Equals, "tcp")
	c.Assert(command.Args, HasLen, 1)
	c.Check(command.Args[0].Name, Equals, "port")
	c.Check(command.Args[0].Value, Equals, "8080")
}
