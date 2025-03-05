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
			Name:     "check-puppet-last-run",
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
			Default:   false,
			Value:     &plugin.IgnoreFailures,
			Usage:     "Ignore Puppet failures, use together with warning, critical about time since last puppet run",
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

	var message string
	var critical_message string
	var warning_message string
	var error_code int
	var critical_code int

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
				critical_message = fmt.Sprintf("Critical: Puppet was last run %s ago\n", formatted_duration(now-last_run))
				critical_code = 2
			}
			if now-last_run > plugin.WarningAge {
				warning_message = warning_message + fmt.Sprintf("Warning: Puppet was last run %s ago\n", formatted_duration(now-last_run))
				error_code = 1
			} else {
				message = message + fmt.Sprintf("Puppet was last run %s ago\n", formatted_duration(now-last_run))
				error_code = 0
			}
		}
	}

	// Check if the summary contains event information and if there are any failures.
	if !plugin.IgnoreFailures {
		if summary.Events != nil {
			if failures, ok := summary.Events["failure"]; ok {
				if failures > 0 {
					critical_message = critical_message + fmt.Sprintf("Last puppet run had failures: %d\n", failures)
					critical_code = 2
				}
			} else {
				warning_message = warning_message + fmt.Sprintf("%s is missing information about the events", plugin.SummaryFile)
				critical_code = 2
			}
		} else {
			fmt.Printf("%s is missing information about the events", plugin.SummaryFile)
			return sensu.CheckStateCritical, nil
		}
	}

	if critical_code == 2 {
		error_code = 2
	}

	switch error_code {
	case 0:
		fmt.Print(message)
		fmt.Printf("Puppet last run was without any failures")
		return sensu.CheckStateOK, nil
	case 1:
		fmt.Print(warning_message)
		return sensu.CheckStateWarning, nil
	case 2:
		fmt.Print(critical_message)
		return sensu.CheckStateCritical, nil
	default:
		return sensu.CheckStateUnknown, nil
	}

}
