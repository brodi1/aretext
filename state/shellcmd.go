package state

import (
	"log"

	"github.com/aretext/aretext/shell"
)

const selectionEnvVar = "SELECTION"

// ScheduleShellCmd schedules a shell command to be executed by the editor.
func ScheduleShellCmd(state *EditorState, shellCmd string) {
	log.Printf("Scheduled shell command: '%s'\n", shellCmd)
	state.scheduledShellCmd = shell.NewCmd(shellCmd, shellCmdEnv(state))
}

func shellCmdEnv(state *EditorState) map[string]string {
	env := make(map[string]string, 1)
	buffer := state.documentBuffer
	r := buffer.SelectedRegion()
	selectionText := copyText(buffer.textTree, r.StartPos, r.EndPos-r.StartPos)
	if len(selectionText) > 0 {
		env[selectionEnvVar] = selectionText
	}
	return env
}
