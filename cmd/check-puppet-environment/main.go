package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/nmollerup/sensu-check-puppet/sensupluginspuppet"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

type Config struct {
	sensu.PluginConfig
	ConfigFile string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "check-puppet-environment",
			Short:    "Check if environment key is set in puppet.conf",
			Keyspace: "",
		},
	}
	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "config-file",
			Argument:  "config-file",
			Shorthand: "c",
			Usage:     "Path to puppet.conf file",
			Value:     &plugin.ConfigFile,
			Default:   sensupluginspuppet.PuppetConfigFile,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	if _, err := os.Stat(plugin.ConfigFile); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - this is OK, just skip the check
			fmt.Printf("OK: File %s not found, skipping check\n", plugin.ConfigFile)
			return sensu.CheckStateOK, nil
		}
		// Other error (e.g., permission denied) - this is critical
		fmt.Printf("Error accessing %s: %v\n", plugin.ConfigFile, err)
		return sensu.CheckStateCritical, nil
	}
	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {
	file, err := os.Open(plugin.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - this is OK
			fmt.Printf("OK: File %s not found, skipping check\n", plugin.ConfigFile)
			return sensu.CheckStateOK, nil
		}
		// Other error (e.g., permission denied) - this is critical
		fmt.Printf("Could not open %s: %v\n", plugin.ConfigFile, err)
		return sensu.CheckStateCritical, nil
	}
	defer func() {
		_ = file.Close()
	}()

	// Match "environment" key in INI format (with optional whitespace)
	envRegex := regexp.MustCompile(`^\s*environment\s*=`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if envRegex.MatchString(line) {
			fmt.Printf("Critical: 'environment' key found in %s\n", plugin.ConfigFile)
			return sensu.CheckStateCritical, nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", plugin.ConfigFile, err)
		return sensu.CheckStateCritical, nil
	}

	fmt.Printf("OK: No 'environment' key found in %s\n", plugin.ConfigFile)
	return sensu.CheckStateOK, nil
}
