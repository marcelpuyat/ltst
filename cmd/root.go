// Copyright © 2016 Marcel Puyat <marcelp@alumni.stanford.edu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os/exec"
	"strings"
)

const BASH_COMPLETION_FILENAME = "latest-completion.sh"

const GEN_AUTOCOMPLETE_FLAG = "gen-autocomplete"
const NUM_RESULTS_FLAG = "num"
const OPEN_FLAG = "open"

const URL_KEY = "url"
const QUERY_KEY = "query"
const NAME_KEY = "name"
const SHORT_DESC_KEY = "shortDescription"
const LONG_DESC_KEY = "longDescription"
const COMMAND_KEY = "command"

var REQUIRED_CONFIG = [...]string{URL_KEY, QUERY_KEY, NAME_KEY, COMMAND_KEY}

const CONFIG_FILE_TYPE = "yaml"
const CONFIG_FILENAME = ".latest"
const CONFIG_DIR = "$HOME"
const CONFIG_FULL_PATH = CONFIG_DIR + "/" + CONFIG_FILENAME + "." + CONFIG_FILE_TYPE

var cfgFile string

// Map of command names to the metadata of the command
var cmdMap map[string]map[interface{}]interface{} = make(map[string]map[interface{}]interface{})

var RootCmd = &cobra.Command{
	Use:   "latest",
	Short: "Get the latest news",
	Long:  "Get the latest news from all of your favorite sites! Fetched in parallel!",
	Run: func(cmd *cobra.Command, args []string) {
		shouldGenerateAutocomplete, err := cmd.Flags().GetBool(GEN_AUTOCOMPLETE_FLAG)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s flag: %s\n", GEN_AUTOCOMPLETE_FLAG, err)
			return
		}

		if shouldGenerateAutocomplete {
			cmd.GenBashCompletionFile(BASH_COMPLETION_FILENAME)
			return
		}

		numResultsPerCommand, err := cmd.Flags().GetInt(NUM_RESULTS_FLAG)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s flag: %s\n", NUM_RESULTS_FLAG, err)
			return
		}

		parallelPrintLatestOfAllCommands(cmd, numResultsPerCommand)
	},
}

func parallelPrintLatestOfAllCommands(rootCmd *cobra.Command, numResultsPerCommand int) {
	results := make(chan string)
	for _, subCmd := range rootCmd.Commands() {
		// Skip over help command
		if subCmd.Name() != "help" {
			entry := cmdMap[subCmd.Use]
			go func(e map[interface{}]interface{}) {
				results <- getNLatest(e[URL_KEY].(string), e[QUERY_KEY].(string),
					numResultsPerCommand, e[NAME_KEY].(string))
			}(entry)
		}
	}

	// -1 is because we are skipping over the help command
	for i := 0; i < len(rootCmd.Commands())-1; i++ {
		fmt.Printf(<-results)
	}
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	initConfig()
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", CONFIG_FULL_PATH))
	RootCmd.Flags().IntP(NUM_RESULTS_FLAG, "n", 1, "Number of results to display")
	RootCmd.Flags().Bool(GEN_AUTOCOMPLETE_FLAG, false, "Generate autocomplete shell script")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigType(CONFIG_FILE_TYPE)
	viper.SetConfigName(CONFIG_FILENAME) // name of config file (without extension)
	viper.AddConfigPath(CONFIG_DIR)      // adding home directory as first search path
	viper.AutomaticEnv()                 // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file %s\n\t%s\n", CONFIG_FULL_PATH, err)
	}
	createCommandsFromEntries(convertRawListToMapsOfStrings(viper.Get("entries").([]interface{})))
}

func convertRawListToMapsOfStrings(list []interface{}) (config []map[interface{}]interface{}) {
	var ret []map[interface{}]interface{}
	for _, e := range list {
		ret = append(ret, e.(map[interface{}]interface{}))
	}
	return ret
}

func createCommandsFromEntries(entries []map[interface{}]interface{}) {
	for _, entry := range entries {
		err := validateEntry(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in config for entry:\n\t%s\n\t%s\n", entry, err)
			continue
		}

		command := entry[COMMAND_KEY].(string)
		cmdMap[command] = entry

		shortDesc := getOrDefault(entry, SHORT_DESC_KEY, entry[NAME_KEY]).(string)
		var newCmd = &cobra.Command{
			Use:   command,
			Short: shortDesc,
			Long:  getOrDefault(entry, LONG_DESC_KEY, shortDesc).(string),
			Run:   printLatestOfCommand,
		}
		RootCmd.AddCommand(newCmd)
		newCmd.Flags().BoolP(OPEN_FLAG, "o", false, fmt.Sprintf("Open %s link", entry[URL_KEY]))
		newCmd.Flags().IntP(NUM_RESULTS_FLAG, "n", 5, "Number of results to display")
	}
}

func validateEntry(entry map[interface{}]interface{}) error {
	for _, cfg := range REQUIRED_CONFIG {
		if entry[cfg] == nil {
			return errors.New(fmt.Sprintf("%s is a required config value", cfg))
		}
	}
	return nil
}

func getOrDefault(entry map[interface{}]interface{}, key string, defaultVal interface{}) interface{} {
	possibleVal := entry[key]
	if possibleVal == nil {
		return defaultVal
	}
	return possibleVal
}

func printLatestOfCommand(cmd *cobra.Command, args []string) {
	openFlag, err := cmd.Flags().GetBool(OPEN_FLAG)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s flag: %s\n", OPEN_FLAG, err)
		return
	}

	numResultsToShow, err := cmd.Flags().GetInt("num")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s flag: %s\n", NUM_RESULTS_FLAG, err)
		return
	}

	entry := cmdMap[cmd.Use]
	url := entry[URL_KEY].(string)
	if openFlag {
		err := exec.Command("open", url).Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running command:\n\topen %s\n\t%s\n", url, err)
		}
		return
	}

	fmt.Println(getNLatest(url, entry[QUERY_KEY].(string), numResultsToShow, entry[NAME_KEY].(string)))
}

func getNLatest(url string, query string, limit int, name string) string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return fmt.Sprintf("Error reaching %s\n\t%s\n", url, err)
	}
	sel := doc.Find(query)
	ret := name + "\n"
	for ix := range sel.Nodes {
		if ix < limit {
			ret += "\t" + strings.Trim(sel.Nodes[ix].FirstChild.Data, " \n\t\r") + "\n"
		}
	}
	return ret
}
