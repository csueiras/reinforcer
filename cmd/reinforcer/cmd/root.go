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
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Version will be set in CI to the current released version
var Version = "0.0.0"

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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

		src, _ := flags.GetString("src")
		sourceTypeName, _ := cmd.Flags().GetString("name")
		outPkg, _ := flags.GetString("outpkg")
		outDir, _ := flags.GetString("outputdir")
		ignoreNoRet, _ := flags.GetBool("ignorenoret")

		if !path.IsAbs(outDir) {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			outDir = path.Join(cwd, path.Clean(outDir))
		} else {
			outDir = path.Clean(outDir)
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}

		l := loader.DefaultLoader()
		_, typ, err := l.Load(src, sourceTypeName)
		if err != nil {
			return err
		}

		code, err := generator.Generate(generator.Config{
			OutPkg:                outPkg,
			IgnoreNoReturnMethods: ignoreNoRet,
			Files: map[string]*generator.FileConfig{
				src: {
					SrcTypeName:   sourceTypeName,
					OutTypeName:   sourceTypeName,
					InterfaceType: typ,
				},
			},
		})
		if err != nil {
			return err
		}

		w := path.Join(outDir, "reinforcer_common.go")
		if err := ioutil.WriteFile(w, []byte(code.Common), 0755); err != nil {
			return fmt.Errorf("failed to write to %s; error=%w", w, err)
		}

		for _, codegen := range code.Files {
			w = path.Join(outDir, toSnakeCase(codegen.TypeName)+".go")
			if err := ioutil.WriteFile(w, []byte(codegen.Contents), 0755); err != nil {
				return fmt.Errorf("failed to write to %s; error=%w", w, err)
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defSrcFile := ""
	if goFile := os.Getenv("GOFILE"); goFile != "" {
		var err error
		defSrcFile, err = os.Getwd()
		if err != nil {
			panic(err)
		}
		defSrcFile = defSrcFile + "/" + goFile
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.reinforcer.yaml)")

	flags := rootCmd.Flags()
	flags.Bool("version", false, "show reinforcer's version")
	flags.String("name", "", "name of interface to generate reinforcer's proxy for")
	flags.String("src", defSrcFile, "source file to scan for the target interface. If unspecified the file pointed by the env variable GOFILE will be used.")
	flags.String("outputdir", "./reinforced", "directory to write the generated code to")
	flags.String("outpkg", "reinforced", "name of generated package")
	flags.Bool("ignorenoret", false, "ignores methods that don't return anything (they won't be wrapped in the middleware). By default they'll be wrapped in a middleware and if the middleware emits an error the call will panic.")
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

// Taken from: https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
