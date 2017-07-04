// 2014 Iain Shigeoka - BSD license (see LICENSE)
package cli_test

import (
	. "github.com/gopackage/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Argument Parsing", func() {

	BeforeEach(func() {
		// Nothing to do yet
	})

	Describe("Option parsing", func() {
		Context("with a single short option flag", func() {
			option := NewOption(nil, "-v", "display version information")
			It("should have a short but no long option", func() {
				Ω(option.Short).Should(Equal("-v"))
				Ω(option.Long).Should(Equal(""))
				Ω(option.Required).Should(BeFalse())
				Ω(option.Optional).Should(Equal(false))
				Ω(option.Bool).Should(Equal(false))
				Ω(option.Description).Should(Equal("display version information"))
			})
		})
		Context("with short and long option flag", func() {
			option := NewOption(nil, "-v, --version", "display version information")
			It("should have both short and long options", func() {
				Ω(option.Short).Should(Equal("-v"))
				Ω(option.Long).Should(Equal("--version"))
				Ω(option.Required).Should(Equal(false))
				Ω(option.Optional).Should(Equal(false))
				Ω(option.Bool).Should(Equal(false))
				Ω(option.Description).Should(Equal("display version information"))

			})
		})
		Context("with a required option parameter", func() {
			option := NewOption(nil, "-c, --config <path>", "set configuration file")
			It("should require an option parameter", func() {
				Ω(option.Short).Should(Equal("-c"))
				Ω(option.Long).Should(Equal("--config"))
				Ω(option.Required).Should(Equal(true))
				Ω(option.Optional).Should(Equal(false))
				Ω(option.Bool).Should(Equal(false))
				Ω(option.Description).Should(Equal("set configuration file"))
			})
		})
		Context("with an optional option parameter", func() {
			option := NewOption(nil, "-c, --config [path]", "set configuration file")
			It("should support the optional parameter", func() {
				Ω(option.Short).Should(Equal("-c"))
				Ω(option.Long).Should(Equal("--config"))
				Ω(option.Required).Should(Equal(false))
				Ω(option.Optional).Should(Equal(true))
				Ω(option.Bool).Should(Equal(false))
				Ω(option.Description).Should(Equal("set configuration file"))
			})
		})
		Context("with an option flag (bool)", func() {
			option := NewOption(nil, "-T, --no-tests", "ignore tests")
			It("should contain a flag option", func() {
				Ω(option.Short).Should(Equal("-T"))
				Ω(option.Long).Should(Equal("--no-tests"))
				Ω(option.Required).Should(Equal(false))
				Ω(option.Optional).Should(Equal(false))
				Ω(option.Bool).Should(Equal(true))
				Ω(option.Description).Should(Equal("ignore tests"))
			})
		})
	})

	Describe("Command parsing", func() {
		Context("with a simple command", func() {
			command := NewCommand(nil, "foo", "bar")
			It("should not expect parameters", func() {
				Ω(command.Command).Should(Equal("foo"))
				Ω(command.Description).Should(Equal("bar"))
			})
		})
		Context("with a required parameter", func() {
			command := NewCommand(nil, "foo <bar>", "a foo bar command")
			It("should require a single parameter", func() {
				Ω(command.Command).Should(Equal("foo"))
				Ω(len(command.Args)).Should(Equal(1))
				Ω(command.Args[0].Name).Should(Equal("bar"))
				Ω(command.Args[0].Required).Should(Equal(true))
				Ω(command.Description).Should(Equal("a foo bar command"))
			})
		})
		Context("with an optional parameter", func() {
			command := NewCommand(nil, "foo [bar]", "a foo bar command")
			It("should support an optional parameter", func() {
				Ω(command.Command).Should(Equal("foo"))
				Ω(len(command.Args)).Should(Equal(1))
				Ω(command.Args[0].Name).Should(Equal("bar"))
				Ω(command.Args[0].Required).Should(Equal(false))
				Ω(command.Description).Should(Equal("a foo bar command"))
			})
		})
	})
	Describe("Normalizing arguments", func() {
		Context("with a simple string", func() {
			normalized := Normalize([]string{"help"})
			It("should return the string", func() {
				Ω(len(normalized)).Should(Equal(1))
				Ω(normalized[0]).Should(Equal("help"))
			})
		})
		Context("with a simple number", func() {
			normalized := Normalize([]string{"8"})
			It("should return the number", func() {
				Ω(len(normalized)).Should(Equal(1))
				Ω(normalized[0]).Should(Equal("8"))
			})
		})
		Context("with a single short option", func() {
			normalized := Normalize([]string{"-v"})
			It("should return the option", func() {
				Ω(len(normalized)).Should(Equal(1))
				Ω(normalized[0]).Should(Equal("-v"))
			})
		})
		Context("with a single long option", func() {
			normalized := Normalize([]string{"--version"})
			It("should return the option", func() {
				Ω(len(normalized)).Should(Equal(1))
				Ω(normalized[0]).Should(Equal("--version"))
			})
		})
		Context("with three short options together", func() {
			normalized := Normalize([]string{"-abc"})
			It("should return the three short options separately", func() {
				Ω(len(normalized)).Should(Equal(3))
				Ω(normalized[0]).Should(Equal("-a"))
				Ω(normalized[1]).Should(Equal("-b"))
				Ω(normalized[2]).Should(Equal("-c"))
			})
		})
		Context("with an option and parameter", func() {
			normalized := Normalize([]string{"--port", "8080"})
			It("should return the long option and parameter separately", func() {
				Ω(len(normalized)).Should(Equal(2))
				Ω(normalized[0]).Should(Equal("--port"))
				Ω(normalized[1]).Should(Equal("8080"))
			})
		})
		Context("with an option and parameter connected with an '='", func() {
			normalized := Normalize([]string{"--port=8080"})
			It("should return the long option and parameter separately", func() {
				Ω(len(normalized)).Should(Equal(2))
				Ω(normalized[0]).Should(Equal("--port"))
				Ω(normalized[1]).Should(Equal("8080"))
			})
		})
	})
	Describe("OptionFor", func() {
		Context("with a single option added", func() {
			program := New()
			It("should retrieve options by short and long flags", func() {
				option := program.OptionFor("-v")
				Ω(option).Should(BeNil())

				program.Option("-v, --version", "display option")

				option = program.OptionFor("-v")
				Ω(option.Name).Should(Equal("version"))

				option = program.OptionFor("--version")
				Ω(option.Name).Should(Equal("version"))

				option = program.OptionFor("-f")
				Ω(option).Should(BeNil())
				option = program.OptionFor("--foo")
				Ω(option).Should(BeNil())
			})
		})
	})
	Describe("ParseOptions", func() {
		Context("with an argument and no configured options", func() {
			program := New()
			args, unknown := program.ParseOptions([]string{"help"})
			It("should leave the argument as-is (left in args[])", func() {
				Ω(len(args)).Should(Equal(1))
				Ω(args[0]).Should(Equal("help"))
				Ω(len(unknown)).Should(Equal(0))
			})
		})
		Context("with a long option and no configured options", func() {
			program := New()
			args, unknown := program.ParseOptions([]string{"--foo"})
			It("should not match the option (add to unknown[] list)", func() {
				Ω(len(args)).Should(Equal(0))
				Ω(len(unknown)).Should(Equal(1))
				Ω(unknown[0]).Should(Equal("--foo"))
			})
		})
		Context("with a long option that matches a configured option", func() {
			program := New()
			program.Option("-f, --foo", "add a foo", "")
			args, unknown := program.ParseOptions([]string{"--foo"})
			option := program.OptionFor("--foo")
			It("should match the option", func() {
				Ω(len(args)).Should(Equal(0))
				Ω(len(unknown)).Should(Equal(0))
				Ω(option.Value).Should(Equal("true"))
			})
		})
	})
	Describe("ParseNormalizedArgs", func() {
		Context("with a configured program", func() {

			program := New()
			program.SetDescription("Device troubleshooting tool")
			program.Option("-v, --verbose", "display verbose information")
			program.Command("tcp <port>", "capture TCP packets on <port>").Option("-h, --host", "host address to bind to")
			program.Topic("path", "setting the path for reading")

			It("should not match unrecognized 'help' command", func() {
				command := program.ParseNormalizedArgs([]string{"help"}, []string{})

				Ω(command).Should(BeNil())
			})
			It("should parse command with required argument", func() {

				command := program.ParseNormalizedArgs([]string{"tcp", "8080"}, []string{})

				Ω(command.Command).Should(Equal("tcp"))
				Ω(len(command.Args)).Should(Equal(1))
				Ω(command.Args[0].Name).Should(Equal("port"))
				Ω(command.Args[0].Value).Should(Equal("8080"))
			})
		})
	})

})
