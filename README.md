[![Go Test](https://github.com/nmollerup/sensu-check-puppet/actions/workflows/test.yml/badge.svg)](https://github.com/nmollerup/sensu-check-puppet/actions/workflows/test.yml) [![goreleaser](https://github.com/nmollerup/sensu-check-puppet/actions/workflows/release.yml/badge.svg)](https://github.com/nmollerup/sensu-check-puppet/actions/workflows/release.yml)

# sensu-check-puppet

Cross platform puppet checks for Sensu compatible monitoring

## Table of Contents

## Overview

## Usage examples

```
  check-puppet-last-run [flags]
  check-puppet-last-run [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --agent-disabled-file string   Path to agent disabled lock file (default "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock")
  -c, --crit-age int                 Age in seconds to be a critical (default 7200)
  -C, --crit-age-disabled int        Age in seconds to crit when agent is disabled (default 7200)
  -d, --disabled-age-limits          Consider disabled age limits, otherwise use main limits (default true)
  -h, --help                         help for check-puppet-last-run
  -i, --ignore-failures              Ignore Puppet failures, use together with warning, critical about time since last puppet run
  -r, --report-restart-failures      Raise alerts if restart failures have happened (default true)
      --summary-file string          Path to summary file for puppet runs (default "/opt/puppetlabs/puppet/public/last_run_summary.yaml")
  -w, --warn-age int                 Age in seconds to be a warning (default 3600)
  -W, --warn-age-disabled int        Age in seconds to warn when agent is disabled (default 3600)

Use "check-puppet-last-run [command] --help" for more information about a command.
```

## Configuration

### Asset registration
