// Command work-manager-prototype is a throwaway terminal harness for issue #64.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dragondad22/codex-starter-kit/internal/prototype/workmanager"
)

func main() {
	state := workmanager.InitialState()
	reader := bufio.NewReader(os.Stdin)
	for {
		render(state)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		key := strings.TrimSpace(line)
		if key == "q" {
			return
		}
		if key == "x" {
			state = workmanager.InitialState()
			continue
		}
		action, ok := actions[key]
		if !ok {
			state.Message = "Unknown key; choose an action listed below."
			continue
		}
		state = workmanager.Reduce(state, action)
	}
}

var actions = map[string]workmanager.Action{
	"p": workmanager.PlanReconciliation,
	"a": workmanager.ApplyNextSuccess,
	"l": workmanager.LoseCreateResponse,
	"u": workmanager.ObserveAmbiguous,
	"r": workmanager.HitRateLimit,
	"o": workmanager.GoOffline,
	"n": workmanager.Reconnect,
	"h": workmanager.RefreshHandshake,
	"m": workmanager.MigrateFieldOption,
	"g": workmanager.AcceptMigration,
	"c": workmanager.CompleteBlocker,
	"s": workmanager.ChangeSource,
	"t": workmanager.ResetRate,
}

func render(state workmanager.State) {
	fmt.Print("\x1b[2J\x1b[H")
	fmt.Println("\x1b[1mPROTOTYPE — deterministic Work Manager reconciliation (#64)\x1b[0m")
	fmt.Println("\x1b[2mQuestion: can desired work and adapter observations converge safely across drift and failure?\x1b[0m")
	fmt.Printf("\n\x1b[1mDisposition:\x1b[0m %s\n\x1b[1mMessage:\x1b[0m %s\n", state.Disposition, state.Message)
	content, _ := json.MarshalIndent(state, "", "  ")
	fmt.Printf("\n%s\n", content)
	fmt.Println("\n\x1b[1mActions\x1b[0m")
	fmt.Println("[p] plan  [a] apply next  [l] lose create response  [u] lookup ambiguous marker")
	fmt.Println("[r] rate limit  [o] offline  [n] reconnect  [h] refresh handshake")
	fmt.Println("[m] migrate option  [g] accept migration  [s] change desired source")
	fmt.Println("[c] complete #64  [t] reset rate window  [x] reset prototype  [q] quit")
}
