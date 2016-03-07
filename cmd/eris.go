package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/ipfs"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

const VERSION = version.VERSION
const dVerMin = version.DVER_MIN

// Defining the root command
var ErisCmd = &cobra.Command{
	Use:   "eris COMMAND [FLAG ...]",
	Short: "The Blockchain Application Platform",
	Long: `Eris is a platform for building, testing, maintaining, and operating
distributed applications with a blockchain backend. Eris makes it easy
and simple to wrangle the dragons of smart contract blockchains.

Made with <3 by Eris Industries.

Complete documentation is available at https://docs.erisindustries.com
` + "\nVersion:\n  " + VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(logger.ErisFormatter{})

		log.SetLevel(log.WarnLevel)
		if do.Verbose {
			log.SetLevel(log.InfoLevel)
		} else if do.Debug {
			log.SetLevel(log.DebugLevel)
		}

		ipfs.IpfsHost = config.GlobalConfig.Config.IpfsHost

		util.DockerConnect(do.Verbose, do.MachineName)

		dockerVersion, err := util.DockerClientVersion()
		if err != nil {
			IfExit(fmt.Errorf("There was an error connecting to your docker daemon.\nCome back after you have resolved and the marmots will be happy to service your blockchain management needs\n\n%v", err))
		}
		marmot := "Come back after you have upgraded and the marmots will be happy to service your blockchain management needs"
		if !util.CompareVersions(dockerVersion, dVerMin) {
			IfExit(fmt.Errorf("Eris requires docker version >= %v\nThe marmots have detected docker version: %v\n%s", dVerMin, dockerVersion, marmot))
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := config.SaveGlobalConfig(config.GlobalConfig.Config)
		if err != nil {
			log.Errorln(err)
		}
	},
}

func Execute() {
	InitializeConfig()
	AddGlobalFlags()
	AddCommands()
	ErisCmd.Execute()
}

// Define the commands
func AddCommands() {
	buildServicesCommand()
	ErisCmd.AddCommand(Services)
	buildChainsCommand()
	ErisCmd.AddCommand(Chains)
	buildPackagesCommand()
	ErisCmd.AddCommand(Packages)
	buildKeysCommand()
	ErisCmd.AddCommand(Keys)
	buildActionsCommand()
	ErisCmd.AddCommand(Actions)

	// TODO
	// buildApplicationsCommand()
	// ErisCmd.AddCommand(Applications)
	buildRemotesCommand()
	ErisCmd.AddCommand(Remotes)

	buildFilesCommand()
	ErisCmd.AddCommand(Files)
	buildDataCommand()
	ErisCmd.AddCommand(Data)
	ErisCmd.AddCommand(ListEverything)

	// TODO
	// buildAgentsCommand()
	// ErisCmd.AddCommand(Agents)

	buildCleanCommand()
	ErisCmd.AddCommand(Clean)
	buildInitCommand()
	ErisCmd.AddCommand(Init)
	buildUpdateCommand()
	ErisCmd.AddCommand(Update)
	buildVerSionCommand()
	ErisCmd.AddCommand(VerSion)

	if runtime.GOOS != "windows" {
		buildManCommand()
		ErisCmd.AddCommand(ManPage)
	}
	ErisCmd.SetHelpCommand(Help)
	ErisCmd.SetHelpTemplate(helpTemplate)
}

// Global Do struct
var do *definitions.Do

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", false, "verbose output")
	ErisCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", false, "debug level output")
	ErisCmd.PersistentFlags().StringVarP(&do.MachineName, "machine", "m", "eris", "machine name for docker-machine that is running VM")
}

func InitializeConfig() {
	var err error
	var out io.Writer
	var erw io.Writer

	do = definitions.NowDo()

	if os.Getenv("ERIS_CLI_WRITER") != "" {
		out, err = os.Open(os.Getenv("ERIS_CLI_WRITER"))
		if err != nil {
			fmt.Printf("Could not open: %s\n", err)
			return
		}
	} else {
		out = os.Stdout
	}

	if os.Getenv("ERIS_CLI_ERROR_WRITER") != "" {
		erw, err = os.Open(os.Getenv("ERIS_CLI_ERROR_WRITER"))
		if err != nil {
			fmt.Printf("Could not open: %s\n", err)
			return
		}
	} else {
		erw = os.Stderr
	}

	config.GlobalConfig, err = config.SetGlobalObject(out, erw)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ArgCheck(num int, comp string, cmd *cobra.Command, args []string) error {
	switch comp {
	case "eq":
		if len(args) != num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of arguments.\nPlease send the marmots %d arguments only.", num)
		}
	case "ge":
		if len(args) < num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of arguments.\nPlease send the marmots at least %d argument(s).", num)
		}
	}
	return nil
}

//restrict flag behaviour when needed (rare but used sometimes)
func FlagCheck(num int, comp string, cmd *cobra.Command, flags []string) error {
	switch comp {
	case "eq":
		if len(flags) != num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of flags.\nPlease send the marmots %d flags only.", num)
		}
	case "ge":
		if len(flags) < num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of flags.\nPlease send the marmots at least %d flag(s).", num)
		}
	}
	return nil
}
