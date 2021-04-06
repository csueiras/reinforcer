//go:generate mockery --all

// MIT License
//
// Copyright (c) 2021 Christian Sueiras
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/generator"
	"github.com/csueiras/reinforcer/internal/generator/executor"
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/csueiras/reinforcer/internal/writer"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

// Version will be set in CI to the current released version
var Version = "0.0.0"
var cfgFile string

// Writer describes the code generator writer
type Writer interface {
	Write(outputDirectory string, generated *generator.Generated) error
}

// Executor describes the code generator executor
type Executor interface {
	Execute(settings *executor.Parameters) (*generator.Generated, error)
}

// DefaultRootCmd creates the default root command with its dependencies wired in
func DefaultRootCmd() *cobra.Command {
	return NewRootCmd(executor.New(loader.DefaultLoader()), writer.Default())
}

// NewRootCmd creates the root command for reinforcer
func NewRootCmd(exec Executor, writ Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "reinforcer",
		Short: "Generates the reinforced middleware code",
		Long: `Reinforcer is a CLI tool that generates code from interfaces that
will automatically inject middleware. Middlewares provide resiliency constructs
such as circuit breaker, retries, timeouts, etc.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			if showVersion, _ := flags.GetBool("version"); showVersion {
				fmt.Println(Version)
				return nil
			}

			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

			debug, _ := flags.GetBool("debug")
			silent, _ := flags.GetBool("silent")

			// Default level for this example is info, unless debug flag is present (or logging is disabled)
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if silent {
				zerolog.SetGlobalLevel(zerolog.Disabled)
			}

			sources, err := flags.GetStringSlice("src")
			if err != nil {
				return err
			}
			sourcePackages, err := flags.GetStringSlice("srcpkg")
			if err != nil {
				return err
			}
			if len(sources)+len(sourcePackages) == 0 {
				goFile := os.Getenv("GOFILE")
				if goFile == "" {
					return fmt.Errorf("no source provided")
				}

				defSrcFile, err := os.Getwd()
				if err != nil {
					return err
				}
				sources = append(sources, path.Join(defSrcFile, goFile))
			}

			targetAll, err := flags.GetBool("targetall")
			if err != nil {
				return err
			}
			targets, err := flags.GetStringSlice("target")
			if err != nil {
				return err
			}
			if len(targets) == 0 && !targetAll {
				return fmt.Errorf("no targets provided")
			}
			outPkg, err := flags.GetString("outpkg")
			if err != nil {
				return err
			}
			outDir, err := flags.GetString("outputdir")
			if err != nil {
				return err
			}
			ignoreNoRet, err := flags.GetBool("ignorenoret")
			if err != nil {
				return err
			}

			gen, err := exec.Execute(&executor.Parameters{
				Sources:               sources,
				SourcePackages:        sourcePackages,
				Targets:               targets,
				TargetsAll:            targetAll,
				OutPkg:                outPkg,
				IgnoreNoReturnMethods: ignoreNoRet,
			})
			if err != nil {
				return fmt.Errorf("failed to generate code; error=%w", err)
			}
			if err := writ.Write(outDir, gen); err != nil {
				return fmt.Errorf("failed to save generated code; error=%w", err)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.reinforcer.yaml)")

	flags := rootCmd.Flags()
	flags.BoolP("version", "v", false, "show reinforcer's version")
	flags.BoolP("debug", "d", false, "enables debug logs")
	flags.BoolP("silent", "q", false, "disables logging. Mutually exclusive with the debug flag.")
	flags.StringSliceP("src", "s", nil, "source files to scan for the target interface or struct. If unspecified the file pointed by the env variable GOFILE will be used.")
	flags.StringSliceP("srcpkg", "k", nil, "source packages to scan for the target interface or struct.")
	flags.StringSliceP("target", "t", []string{}, "name of target type or regex to match interface or struct names with")
	flags.BoolP("targetall", "a", false, "codegen for all exported interfaces/structs discovered. This option is mutually exclusive with the target option.")
	flags.StringP("outputdir", "o", "./reinforced", "directory to write the generated code to")
	flags.StringP("outpkg", "p", "reinforced", "name of generated package")
	flags.BoolP("ignorenoret", "i", false, "ignores methods that don't return anything (they won't be wrapped in the middleware). By default they'll be wrapped in a middleware and if the middleware emits an error the call will panic.")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".reinforcer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".reinforcer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
