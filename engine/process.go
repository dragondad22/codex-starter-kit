package engine

import (
	"context"
	"os"
	"os/exec"
)

var processEnvironmentAllowlist = []string{
	"PATH", "PATHEXT", "SYSTEMROOT", "COMSPEC", "HOME", "USERPROFILE",
	"HOMEDRIVE", "HOMEPATH", "TEMP", "TMP", "TMPDIR",
}

func structuredGitCommand(ctx context.Context, root string, arguments ...string) *exec.Cmd {
	commandArguments := append([]string{
		"-c", "core.fsmonitor=false",
		"-c", "core.hooksPath=" + os.DevNull,
		"-C", root,
	}, arguments...)
	command := exec.CommandContext(ctx, "git", commandArguments...)
	environment := make([]string, 0, len(processEnvironmentAllowlist)+5)
	for _, key := range processEnvironmentAllowlist {
		if value, present := os.LookupEnv(key); present {
			environment = append(environment, key+"="+value)
		}
	}
	command.Env = append(environment,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_TERMINAL_PROMPT=0",
		"LC_ALL=C",
	)
	return command
}
