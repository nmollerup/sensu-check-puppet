package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nmollerup/sensu-check-puppet/sensupluginspuppet"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"gopkg.in/yaml.v2"
)

type Config struct {
	sensu.PluginConfig
	SummaryFile string
	Scheme      string
}

var (
	defaultScheme = func() string {
		hostname, err := os.Hostname()
		if err != nil {
			return "unknown.puppet"
		}
		return hostname + ".puppet"
	}()

	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "metrics-puppet-run",
			Short:    "Output Puppet run metrics in Graphite plaintext format",
			Keyspace: "",
		},
	}
	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "summary-file",
			Argument:  "summary-file",
			Shorthand: "p",
			Usage:     "Path to last_run_summary.yaml",
			Value:     &plugin.SummaryFile,
			Default:   sensupluginspuppet.SummaryFile,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "scheme",
			Argument:  "scheme",
			Shorthand: "s",
			Usage:     "Metric naming scheme prefix",
			Value:     &plugin.Scheme,
			Default:   defaultScheme,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	if _, err := os.Stat(plugin.SummaryFile); os.IsNotExist(err) {
		fmt.Printf("File %s not found\n", plugin.SummaryFile)
		return sensu.CheckStateUnknown, nil
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	type Summary struct {
		Resources map[string]float64 `yaml:"resources"`
		Time      map[string]float64 `yaml:"time"`
		Changes   map[string]float64 `yaml:"changes"`
		Events    map[string]float64 `yaml:"events"`
	}

	data, err := os.ReadFile(plugin.SummaryFile)
	if err != nil {
		fmt.Printf("Could not process %s: %v\n", plugin.SummaryFile, err)
		return sensu.CheckStateUnknown, nil
	}

	var summary Summary
	if err := yaml.Unmarshal(data, &summary); err != nil {
		fmt.Printf("Could not process %s: %v\n", plugin.SummaryFile, err)
		return sensu.CheckStateUnknown, nil
	}

	now := time.Now().Unix()
	sections := map[string]map[string]float64{
		"resources": summary.Resources,
		"time":      summary.Time,
		"changes":   summary.Changes,
		"events":    summary.Events,
	}

	for _, section := range []string{"resources", "time", "changes", "events"} {
		for key, value := range sections[section] {
			fmt.Printf("%s.%s.%s %g %d\n", plugin.Scheme, section, key, value, now)
		}
	}

	return sensu.CheckStateOK, nil
}
