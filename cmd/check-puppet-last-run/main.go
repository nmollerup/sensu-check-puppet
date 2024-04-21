package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nmollerup/sensu-check-puppet/sensupluginspuppet"
	"gopkg.in/yaml.v2"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

type Config struct {
	sensu.PluginConfig
	AgentDisabledFile     string `json:"agent_disabled_file"`
	SummaryFile           string
	WarningAge            int
	CriticalAge           int
	CriticalAgeDisabled   int
	WarningAgeDisabled    int
	DisabledAgeLimits     bool
	ReportRestartFailures bool
	IgnoreFailures        bool
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
		&sensu.PluginConfigOption[int]{
			Path:      "warn-age",
			Argument:  "warn-age",
			Shorthand: "w",
			Default:   3600,
			Value:     &plugin.WarningAge,
			Usage:     "Age in seconds to be a warning",
		},
		&sensu.PluginConfigOption[int]{
			Path:      "crit-age",
			Argument:  "crit-age",
			Shorthand: "c",
			Default:   7200,
			Value:     &plugin.CriticalAge,
			Usage:     "Age in seconds to be a critical",
		},
		&sensu.PluginConfigOption[int]{
			Path:      "crit-age-disabled",
			Argument:  "crit-age-disabled",
			Shorthand: "C",
			Default:   7200,
			Value:     &plugin.CriticalAgeDisabled,
			Usage:     "Age in seconds to crit when agent is disabled",
		},
		&sensu.PluginConfigOption[int]{
			Path:      "",
			Argument:  "warn-age-disabled",
			Shorthand: "W",
			Default:   3600,
			Value:     &plugin.WarningAgeDisabled,
			Usage:     "Age in seconds to warn when agent is disabled",
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Argument:  "agent-disabled-file",
			Shorthand: "a",
			Default:   sensupluginspuppet.AgentDisabledFile,
			Value:     &plugin.AgentDisabledFile,
			Usage:     "Path to agent disabled lock file",
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "",
			Argument:  "disabled-age-limits",
			Shorthand: "d",
			Default:   true,
			Value:     &plugin.DisabledAgeLimits,
			Usage:     "Consider disabled age limits, otherwise use main limits",
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "",
			Argument:  "report-restart-failures",
			Shorthand: "r",
			Default:   true,
			Value:     &plugin.ReportRestartFailures,
			Usage:     "Raise alerts if restart failures have happened",
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "",
			Argument:  "ignore-failures",
			Shorthand: "i",
			Default:   true,
			Value:     &plugin.IgnoreFailures,
			Usage:     "Ignore Puppet failures",
		},
	}
)

func formatted_duration(seconds int) string {
	hours := seconds / (60 * 60)
	minutes := seconds / 60 % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

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
	now := int(time.Now().Unix())

	type Summary struct {
		Events    map[string]int `yaml:"events"`
		Time      map[string]int `yaml:"time"`
		Resources map[string]int `yaml:"resources"`
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

	// Check if the last run time is available and if it exceeds the critical age threshold.
	if summary.Time != nil {
		if last_run, ok := summary.Time["last_run"]; ok {
			if now-last_run > plugin.CriticalAge {
				fmt.Printf("Puppet last run %s ago\n", formatted_duration(now-last_run))
				return sensu.CheckStateCritical, nil
			}
		}
	}

	// Check if the summary contains event information and if there are any failures.
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

	if !plugin.IgnoreFailures {
		failures := summary.Events["failure"]
		if failures > 0 {
			fmt.Printf("Puppet last run with %d failures", failures)
			return sensu.CheckStateCritical, nil
		}
	}
	return sensu.CheckStateOK, nil

}
