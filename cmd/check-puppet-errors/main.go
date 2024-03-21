package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nmollerup/sensu-check-puppet/sensupluginspuppet"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"gopkg.in/yaml.v2"
)

type Config struct {
	sensu.PluginConfig
	AgentDisabledFile string `json:"agent_disabled_file"`
	SummaryFile       string
}

type DisabledMessage struct {
	DisabledMessage string `json:"disabled_message"`
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-puppet-errors",
			Short:    "",
			Keyspace: "",
		},
	}
	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:     "summary-file",
			Argument: "summary-file",
			Usage:    "Path to summary file for puppet runs",
			Value:    &plugin.SummaryFile,
			Default:  sensupluginspuppet.SummaryFile,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	// Check that summary file can be opened
	if _, err := os.Stat(plugin.SummaryFile); os.IsNotExist(err) {
		fmt.Printf("File %s not found\n", plugin.SummaryFile)
		return sensu.CheckStateCritical, nil
	}

	// Check if the agent has been disabled
	if _, err := os.Stat(plugin.AgentDisabledFile); os.IsNotExist(err) {
	} else {
		data, err := os.ReadFile(plugin.AgentDisabledFile)
		if err != nil {
			fmt.Printf("Error reading agent lockfile: %s\n", err)
			return sensu.CheckStateCritical, nil
		}
		var disabledMessage DisabledMessage
		err = json.Unmarshal(data, &disabledMessage)
		if err != nil {
			fmt.Printf("Error parsing JSON: %s\n", err)
			return sensu.CheckStateCritical, nil
		}
		// return the agent disabled message
		message := fmt.Sprintf("Lockfile exists. Disabled reason: %s", disabledMessage.DisabledMessage)

		fmt.Println(message)
		return sensu.CheckStateCritical, nil

	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {

	type Summary struct {
		Events map[string]int `yaml:"events"`
	}

	data, err := os.ReadFile(plugin.SummaryFile)
	if err != nil {
		fmt.Printf("Could not process %s: %v", plugin.SummaryFile, err)
		return sensu.CheckStateCritical, nil
	}

	var summary Summary
	err = yaml.Unmarshal(data, &summary)
	if err != nil {
		fmt.Printf("Could not process %s: %v", plugin.SummaryFile, err)
		return sensu.CheckStateCritical, nil
	}

	if summary.Events != nil {
		if failures, ok := summary.Events["failure"]; ok {
			fmt.Printf("Failures: %d\n", failures)
			if failures > 0 {
				return sensu.CheckStateCritical, nil
			}
		} else {
			fmt.Printf("%s is missing information about the events", plugin.SummaryFile)
			return sensu.CheckStateCritical, nil
		}
	} else {
		fmt.Printf("%s is missing information about the events", plugin.SummaryFile)
		return sensu.CheckStateCritical, nil
	}
	return sensu.CheckStateOK, nil

}
