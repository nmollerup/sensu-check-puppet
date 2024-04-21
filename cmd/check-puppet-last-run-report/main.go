package main

import (
	"fmt"
	"os"

	"github.com/nmollerup/sensu-check-puppet/sensupluginspuppet"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

type Config struct {
	sensu.PluginConfig
	AgentDisabledFile string `json:"agent_disabled_file"`
	SummaryFile       string
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
	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	return sensu.CheckStateOK, nil
}
