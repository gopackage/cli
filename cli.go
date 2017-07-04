package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Program captures all the information about command line options and
// command syntax. The Program is configured, then Parse() is called to
// trigger program execution.
type Program struct {
	Version        string
	Name           string
	Description    string
	Exe            string
	Execs          map[string]string
	Args           []string
	Commands       map[string]*Command
	Options        map[string]*Option
	Topics         map[string]*Topic
	RunningCommand *exec.Cmd

	// Terminal attached to this program
	Terminal *Terminal
}

// New creates a new command line program.
func New() *Program {
	program := &Program{Commands: map[string]*Command{}, Options: map[string]*Option{}, Topics: map[string]*Topic{}}
	program.Terminal = NewTerminal(program)
	return program
}

// SetName configures the program "pretty name" used in version reporting.
func (p *Program) SetName(name string) *Program {
	p.Name = name
	return p
}

// SetDescription sets the short program description for help summary.
func (p *Program) SetDescription(description string) *Program {
	p.Description = description
	return p
}

// SetVersion sets the program version to `version`.
//
// This method auto-registers the "version" command
// which will print the version number when passed.
func (p *Program) SetVersion(version string, command ...string) *Program {
	p.Version = version
	cmd := "version"
	desc := "output version number"
	body := "Displays the program's version number."
	switch len(command) {
	case 0:
	case 1:
		cmd = command[0]
		fallthrough
	case 2:
		desc = command[1]
	case 3:
		body = command[2]
	}
	versionCommand := p.Command(cmd, desc)
	versionCommand.SetBody(body)
	versionCommand.SetAction(func(program *Program, command *Command, unknownArgs []string) {
		name := p.Exe
		if p.Name != "" {
			name = p.Name
		}
		fmt.Printf("%s -- v %s\n\n", name, p.Version)
	})
	return p
}

// Option adds an option with help message information.
func (p *Program) Option(flags, description string, defaultValue ...string) *Program {
	o := NewOption(p, flags, description, defaultValue...)
	p.Options[flags] = o
	return p
}

// Command adds a command/mode for the program to execute. Use the command
// `*` to register a default command (otherwise the default will be the help command).
func (p *Program) Command(command, description string) *Command {

	c := NewCommand(p, command, description)
	p.Commands[c.Command] = c

	return c
}

// Topic adds a help topic to the program (information without a corresponding)
// program execution.
func (p *Program) Topic(topic, description string) *Topic {
	t := &Topic{Program: p, Topic: topic, Description: description}
	p.Topics[topic] = t
	return t
}

// Parse begins command line argument parsing and returns the command that
// the user selected for execution. If no command was selected the command
// will be nil and Help() will be displayed.
func (p *Program) Parse() *Command {
	return p.ParseArgs(os.Args)
}

// ParseArgs parses the provided list of command line arguments instead of
// automatically pulling them from `os.Args`. Useful for testing.
// See Parse for details.
func (p *Program) ParseArgs(argv []string) *Command {
	// Add implicit help command if there isn't one set
	if _, ok := p.Commands["help"]; !ok {
		helpCommand := NewCommand(p, "help [cmd]", "display help for [cmd]")
		helpCommand.SetAction(HelpAction)
		p.Commands["help"] = helpCommand
	}

	// Binary name
	p.Exe = path.Base(argv[0])

	// process argv
	args, unknown := p.ParseOptions(Normalize(argv[1:]))
	p.Args = args

	result := p.ParseNormalizedArgs(p.Args, unknown)

	// executable sub-commands
	if result == nil {
		// Run the default command actions
		if help, ok := p.Commands["*"]; ok {
			if help.Action != nil {
				help.Action(p, help, unknown)
			}
		} else {
			p.Help()
		}
	} else {
		if _, ok := p.Execs[result.Command]; ok {
			return p.executeSubCommand(argv, args, unknown)
		}
	}

	return result
}

// Execute a sub-command executable.
func (p *Program) executeSubCommand(argv, args, unknown []string) (cmd *Command) {
	args = append(args, unknown...)

	if len(args) == 0 {
		p.Help()
	}

	if "help" == args[0] && 1 == len(args) {
		p.Help()
	}

	// <cmd> --help
	if "help" == args[0] {
		args[0] = args[1]
		args[1] = "--help"
	}

	// executable
	dir := path.Dir(argv[1])
	bin := path.Base(argv[1]) + "-" + args[0]

	// check for ./<bin> first
	local := path.Join(dir, bin)

	// run it
	args = args[1:]
	proc := exec.Command(local, args...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	if err := proc.Run(); err != nil {
		/*
		   	if (err.code == "ENOENT") {
		     		console.error("\n  %s(1) does not exist, try --help\n", bin)
		   	} else if (err.code == "EACCES") {
		     		console.error("\n  %s(1) not executable. try chmod or run with root\n", bin)
		   	}
		*/
		// Print the error for now
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	p.RunningCommand = proc
	return
}

// Normalize `args`, splitting joined short flags. For example
// the arg "-abc" is equivalent to "-a -b -c".
// This also normalizes equal sign and splits "--abc=def" into "--abc def".
func Normalize(args []string) (normalized []string) {

	for _, arg := range args {
		if len(arg) > 1 && "-" == arg[0:1] && "-" != arg[1:2] {
			for _, c := range arg[1:] {
				flag := "-" + string(c)
				normalized = append(normalized, flag)
			}
		} else if len(arg) > 1 && "--" == arg[0:2] && strings.Contains(arg, "=") {
			index := strings.Index(arg, "=")
			normalized = append(normalized, arg[0:index])
			normalized = append(normalized, arg[index+1:])
		} else {
			normalized = append(normalized, arg)
		}
	}
	return
}

// ParseNormalizedArgs parses command line `args` and selects a command based
// on program settings and the arguments.
func (p *Program) ParseNormalizedArgs(args, unknown []string) (command *Command) {
	if len(args) > 0 {
		name := args[0]
		var ok bool
		if command, ok = p.Commands[name]; ok {
		} else if command, ok = p.Commands["*"]; ok {
		} else {
			p.outputHelpIfNecessary(name, unknown)
			return
		}
	} else {
		p.outputHelpIfNecessary("", unknown)

		// If there were no args and we have unknown options,
		// then they are extraneous and we need to error.
		if len(unknown) > 0 {
			p.unknownOption(unknown[0])
			return
		}
	}
	// Set up the remaining command args
	if command != nil {
		args = args[1:]
		for _, arg := range command.Args {
			if len(args) > 0 {
				arg.Value = args[0]
				args = args[1:]
			} else {
				// We ran out of arguments, check if we are missing a requirement
				if arg.Required {
					p.missingArgument(arg.Name)
				}
			}
		}
		if command.Action != nil {
			command.Action(p, command, unknown)
		}
	}
	return
}

// OptionFor returns an option matching `arg` if any.
func (p *Program) OptionFor(arg string) *Option {
	for _, option := range p.Options {
		if option.Short == arg || option.Long == arg {
			return option
		}
	}
	return nil
}

// ParseOptions parses options from `argv` returning `argv` void of these options.
func (p *Program) ParseOptions(argv []string) (args, unknownOptions []string) {
	literal := false

	// parse options
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		// literal args after --
		if "--" == arg {
			literal = true
			continue
		}
		if literal {
			args = append(args, arg)
			continue
		}
		// find matching Option
		option := p.OptionFor(arg)

		// option is defined
		if option != nil {
			if option.Required { // requires arg
				i++
				if len(argv) < i {
					p.optionMissingArgument(option, "")
				}
				arg = argv[i]
				if "-" == arg[0:1] && "-" != arg {
					p.optionMissingArgument(option, arg)
				}
				option.Value = arg
			} else if option.Optional { // optional arg
				if len(argv) > i+1 {
					arg = argv[i+1]
					if "" == arg || ("-" == arg[0:1] && "-" != arg) {
						option.Value = "true"
					} else {
						i++
						option.Value = arg
					}
				} else {
					option.Value = "true"
				}
			} else {
				option.Value = "true"
			}
			continue
		}
		// looks like an option
		if len(arg) > 1 && "-" == arg[0:1] {
			unknownOptions = append(unknownOptions, arg)

			// If the next argument looks like it might be an argument for this
			// option, we pass it on. If it isn't, then it'll simply be ignored
			if len(argv) > i+1 && "-" != argv[i+1][0:1] {
				i++
				unknownOptions = append(unknownOptions, argv[i])
			}
			continue
		}
		// arg
		args = append(args, arg)
	}
	return
}

// Argument `name` is missing.
func (p *Program) missingArgument(name string) {
	fmt.Fprintf(os.Stderr, "\n  error: missing required argument `%s`\n\n", name)
	os.Exit(1)
}

// `Option` is missing an argument, but received `flag` or nothing.
func (p *Program) optionMissingArgument(option *Option, flag string) {
	if flag != "" {
		fmt.Fprintf(os.Stderr, "\n  error: option `%s` argument missing, got `%s`\n\n", option.Flags, flag)
	} else {
		fmt.Fprintf(os.Stderr, "\n  error: option `%s` argument missing\n\n", option.Flags)
	}
	os.Exit(1)
}

// Unknown command argument
func (p *Program) unknownArgument(cmd, arg string) {
	fmt.Fprintf(os.Stderr, "\n  error: command `%s` has unknown argument `%s`\n\n", cmd, arg)
	os.Exit(1)
}

// Unknown option `flag`.
func (p *Program) unknownOption(flag string) {
	fmt.Fprintf(os.Stderr, "\n  error: unknown option `%s`\n\n", flag)
	os.Exit(1)
}

// outputHelpIfNecessary but only if necessary
func (p *Program) outputHelpIfNecessary(cmd string, options []string) {
	for _, option := range options {
		if option == "--help" || option == "-h" {
			p.Help()
		}
	}
}

// PrintHelp displays help message (does not exit).
func (p *Program) PrintHelp() {
	if help, ok := p.Commands["help"]; ok {
		if help.Action != nil {
			help.Action(p, help, []string{})
		}
	}
}

// Help displays help message and exits.
func (p *Program) Help() {
	p.PrintHelp()
	os.Exit(0)
}

// -----------------------------------------------------------------------

// NewCommand creates a new command for a given program. Use the command string
// `*` to indicate a default command that is run when no command is specied.
// By default the default command is the help command.
func NewCommand(program *Program, command, description string) *Command {
	c := &Command{Program: program, Description: description}
	c.Flags = command
	if len(command) > 0 {
		args := regexp.MustCompile(` +`).Split(command, -1)
		c.Command = args[0]
		c.parseExpectedArgs(args[0:])
	}
	return c
}

// CommandAction is implemented by any function wanting to be called when
// a command is selected on the command line.
type CommandAction func(program *Program, command *Command, unknownArgs []string)

// Command captures information about a cli command that the user wishes
// to select/execute. Each command can have its own set of unique options.
// When a command is selected as part of Program.Parse() any CommandActions
// associated with the command are executed in the main thread.
type Command struct {
	Program     *Program
	Command     string
	Flags       string
	Description string
	Body        string
	Args        []*Arg
	Options     []*Option
	Action      CommandAction
}

// Option captures information about a cli option (denoted by a `-` or long `--`
// prefix). We accept flags that look like `--flag=foo` or `-f foo`.
func (c *Command) Option(flags, description string, defaultValue ...string) *Command {
	o := NewOption(c.Program, flags, description, defaultValue...)
	c.Options = append(c.Options, o)
	return c
}

// OptionFor returns an option matching `name` if any.
func (c *Command) OptionFor(name string) *Option {
	for _, option := range c.Options {
		if option.Short == name || option.Long == name {
			return option
		}
	}
	return nil
}

// ArgFor returns an arg matching `name` if any.
func (c *Command) ArgFor(name string) *Arg {
	for _, arg := range c.Args {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}

func (c *Command) parseExpectedArgs(args []string) {
	for _, arg := range args {
		if len(arg) > 0 {
			switch arg[0:1] {
			case "<":
				// No optional arguments before required arguments
				for _, prev := range c.Args {
					if !prev.Required {
						fmt.Fprintf(os.Stderr, "\n  error: required argument `%s` not allowed after optional argument `%s`", arg, prev.Name)
						os.Exit(1)
					}
				}
				c.Args = append(c.Args, &Arg{Required: true, Name: arg[1 : len(arg)-1]})
			case "[":
				c.Args = append(c.Args, &Arg{Required: false, Name: arg[1 : len(arg)-1]})
			}
		}
	}
}

// SetBody configures the body text of a command help description.
func (c *Command) SetBody(body string) *Command {
	c.Body = body
	return c
}

// SetAction sets the action associated with the command.
func (c *Command) SetAction(action CommandAction) *Command {
	c.Action = action
	return c
}

// Arg captures command line arguments that the program expects.
type Arg struct {
	Required bool
	Name     string
	Value    string
}

// IntValue retrieves the current value of the Arg as an int -
// if the value isn't parseable, the provided default is returned
func (a *Arg) IntValue(defaultValue int) int {
	if a.Value == "" {
		return defaultValue
	}
	intVal, err := strconv.ParseInt(a.Value, 0, 0)
	if err != nil {
		return defaultValue
	}
	return (int)(intVal)
}

// -----------------------------------------------------------------------

// Option represents a command line option with both short and long flag formats
// supported.
type Option struct {
	Program     *Program
	Flags       string
	Required    bool
	Optional    bool
	Bool        bool
	Short       string
	Long        string
	Name        string
	Description string
	Value       string
	Default     string
}

// NewOption creates a new option.
func NewOption(program *Program, flags, description string, defaultValue ...string) *Option {
	option := &Option{Program: program}
	option.Flags = flags
	option.Description = description
	option.Required = strings.Contains(flags, "<")
	option.Optional = strings.Contains(flags, "[")
	option.Bool = strings.Contains(flags, "-no-")
	options := regexp.MustCompile(`[ ,|]+`).Split(flags, -1)
	option.Short = options[0]
	if len(options) > 1 {
		option.Long = options[1]
		option.Name = strings.Replace(strings.Replace(option.Long, "--", "", -1), "no-", "", -1)
	}
	if len(defaultValue) == 1 {
		option.Default = defaultValue[0]
	}
	return option
}

// -----------------------------------------------------------------------

// Topic represents a help topic that has no corresponding command (only used
// by the `help` built in command).
type Topic struct {
	Program     *Program
	Description string
	Topic       string
	Body        string
}

// SetDescription sets the help topic description.
func (t *Topic) SetDescription(description string) *Topic {
	t.Description = description
	return t
}

// SetBody sets the help topic body text.
func (t *Topic) SetBody(body string) *Topic {
	t.Body = body
	return t
}

// -----------------------------------------------------------------------

// HelpAction is a default action used by cli to print out the standard
// Help() message. The action can be replaced by a user-supplied implementation
// to override the default behavior/format.
func HelpAction(program *Program, command *Command, _ []string) {
	// Print help - we look it here are any arguments (command or topics) and print those,
	// otherwise, we print the main usage information
	if command != nil {
		cmd := command.Args[0].Value

		// Search commands for a match
		helpCommand := program.Commands[cmd]
		if helpCommand != nil && helpCommand.Command != "" {
			fmt.Print("Usage: ", program.Exe)
			if len(helpCommand.Options) > 0 {
				fmt.Print(" [options]")
			}
			fmt.Println(" " + helpCommand.Flags)
			fmt.Println()
			if helpCommand.Body != "" {
				fmt.Println(helpCommand.Body)
			} else {
				fmt.Println(helpCommand.Description)
			}
			return
		}
		// Search topics for a match
		helpTopic := program.Topics[cmd]
		if helpTopic != nil {
			fmt.Println(helpTopic.Topic)
			line := make([]string, len(helpTopic.Topic))
			for i := range helpTopic.Topic {
				line[i] = "="
			}
			fmt.Println(line)
			fmt.Println()
			if helpTopic.Body != "" {
				fmt.Println(helpTopic.Body)
			} else {
				fmt.Println(helpTopic.Description)
			}
			return
		}
	}
	HelpPrinter(program)
}

// HelpPrinter is the default help printing function
func HelpPrinter(p *Program) {
	defaultCommand, hasDefaultCommand := p.Commands["*"]

	if p.Description != "" {
		fmt.Println(p.Description)
		fmt.Println()
	}

	fmt.Print("Usage: ", p.Exe)
	if len(p.Options) > 0 {
		fmt.Print(" [options]")
	}
	if len(p.Commands) > 0 {
		if hasDefaultCommand {
			fmt.Print(" [command]")
		} else {
			fmt.Print(" <command>")
		}
	}
	fmt.Println()
	fmt.Println()

	// We right pad by spaces all items from descriptions to create a nice lined up list of descriptions
	// To do that, we iterate through all the items and find the longest and pad it by 3 spaces
	padding := "     "
	spacing := "                                                                                        "
	spacer := 3

	columnSize := spacer
	for _, opt := range p.Options {
		if columnSize < len(opt.Flags)+spacer {
			columnSize = len(opt.Flags) + spacer
		}
	}
	if columnSize > len(spacing) {
		columnSize = len(spacing)
	}

	if len(p.Options) > 0 {
		fmt.Println("Global options are:")
		fmt.Println()
		for _, option := range p.Options {
			fmt.Print(padding)
			fmt.Print(option.Flags)
			if len(option.Flags) < columnSize {
				fmt.Print(spacing[0 : columnSize-len(option.Flags)])
			}
			if option.Default != "" {
				fmt.Printf("%s (defaults to %v)\n", option.Description, option.Default)
			} else {
				fmt.Println(option.Description)
			}
		}
		fmt.Println()
	}

	columnSize = spacer
	for _, cmd := range p.Commands {
		if columnSize < len(cmd.Flags)+spacer {
			columnSize = len(cmd.Flags) + spacer
		}
	}
	for _, topic := range p.Topics {
		if columnSize < len(topic.Topic)+spacer {
			columnSize = len(topic.Topic) + spacer
		}
	}

	if len(p.Commands) > 0 {
		fmt.Println("The commands are:")
		fmt.Println()
		for _, command := range p.Commands {
			if command.Flags == "*" {
				// Skip default command in command list - we display it at the bottom
				continue
			}
			fmt.Print(padding)
			fmt.Print(command.Flags)
			if len(command.Flags) < columnSize {
				fmt.Print(spacing[0 : columnSize-len(command.Flags)])
			}
			fmt.Println(command.Description)
		}
		fmt.Println()
		fmt.Println("Use \"" + p.Exe + " help [command]\" for more information about a command.")
		fmt.Println()
	}
	if len(p.Topics) > 0 {
		fmt.Println("Additional help topics:")
		fmt.Println()
		for _, topic := range p.Topics {
			fmt.Print(padding)
			fmt.Print(topic.Topic)
			if len(topic.Topic) < columnSize {
				fmt.Print(spacing[0 : columnSize-len(topic.Topic)])
			}
			fmt.Println(topic.Description)
		}
		fmt.Println()
		fmt.Println("Use \"" + p.Exe + " help [topic]\" for more information about that topic.")
		fmt.Println()
	}

	if hasDefaultCommand {
		// We have a default command
		fmt.Println("Default command:", defaultCommand.Description)
		if defaultCommand.Body != "" {
			fmt.Println(defaultCommand.Body)
		}
		fmt.Println()
	}
}
