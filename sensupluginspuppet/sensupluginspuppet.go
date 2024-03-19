package sensupluginspuppet

import (
	"runtime"
)

var (
	SummaryFile       = getPath("C:/ProgramData/PuppetLabs", "/opt/puppetlabs") + "/puppet/public/last_run_summary.yaml"
	ReportFile        = getPath("C:/ProgramData/PuppetLabs", "/opt/puppetlabs") + "/puppet/cache/state/last_run_report.yaml"
	AgentDisabledFile = getPath("C:/ProgramData/PuppetLabs", "/opt/puppetlabs") + "/puppet/cache/state/agent_disabled.lock"
)

func getPath(windowsPath, unixPath string) string {
	if runtime.GOOS == "windows" {
		return windowsPath
	}
	return unixPath
}
