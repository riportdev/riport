//go:build !windows
// +build !windows

package system

import (
	chshare "github.com/riportdev/riport/share"
)

func getCmdBuildTestcases() []cmdBuildTestCase {
	return []cmdBuildTestCase{
		{
			name:       "empty",
			wantCmdStr: "/bin/sh /script.sh",
			command:    "/script.sh",
		},
		{
			name:        "non empty sh",
			interpreter: chshare.UnixShell,
			wantCmdStr:  "/bin/sh /script.sh",
			command:     "/script.sh",
		},
		{
			name:           "hasShebang, interpreter empty",
			interpreter:    "",
			boolHasShebang: true,
			wantCmdStr:     "/script.sh",
			command:        "/script.sh",
		},
		{
			name:           "hasShebang, interpreter not empty",
			interpreter:    chshare.UnixShell,
			wantCmdStr:     "/script.sh",
			boolHasShebang: true,
			command:        "/script.sh",
		},
		{
			name:         "tacoscript interpreter",
			interpreter:  chshare.Tacoscript,
			partialMatch: true,
			wantCmdStr:   "tacoscript /script.sh",
			command:      "/script.sh",
		},
		{
			name:               "interpreter aliases",
			interpreter:        "taco",
			wantCmdStr:         "/non-standard-interpreter /script.sh",
			interpreterAliases: map[string]string{"taco": "/non-standard-interpreter"},
			command:            "/script.sh",
		},
		{
			name:        "interpreter full path",
			interpreter: `/non-standard-interpreter`,
			wantCmdStr:  "/non-standard-interpreter /script.sh",
			command:     "/script.sh",
		},
	}
}
