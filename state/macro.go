package state

import "log"

// MacroAction is a transformation of editor state that can be recorded and replayed.
type MacroAction func(*EditorState)

// MacroState stores recorded macros.
// The "last action" macro is used to repeat the last logical action
// (using the "." command in normal mode).
type MacroState struct {
	lastActions            []MacroAction
	isRecordingUserMacro   bool
	isReplayingUserMacro   bool
	userMacroActions       []MacroAction
	stagedUserMacroActions []MacroAction
}

// AddToLastActionMacro adds an action to the "last action" macro.
func AddToLastActionMacro(s *EditorState, action MacroAction) {
	s.macroState.lastActions = append(s.macroState.lastActions, action)
}

// ClearLastActionMacro resets the "last action" macro.
func ClearLastActionMacro(s *EditorState) {
	s.macroState.lastActions = nil
}

// ReplayLastActionMacro executes the actions recorded in the "last action" macro.
func ReplayLastActionMacro(s *EditorState, count uint64) {
	for i := uint64(0); i < count; i++ {
		for _, action := range s.macroState.lastActions {
			action(s)
		}
	}
}

// ToggleUserMacroRecording stops/starts recording a user-defined macro.
// If recording stops before any actions have been recorded, the previously-recorded
// macro will be preserved.
func ToggleUserMacroRecording(s *EditorState) {
	m := &s.macroState
	if m.isRecordingUserMacro {
		log.Printf("Stopped recording user macro\n")
		m.isRecordingUserMacro = false

		if len(m.stagedUserMacroActions) == 0 {
			// The user probably started recording by mistake and wouldn't
			// want to lose the previously-recorded macro.
			SetStatusMsg(s, StatusMsg{
				Style: StatusMsgStyleSuccess,
				Text:  "Cancelled macro recording",
			})
			return
		}

		m.userMacroActions = m.stagedUserMacroActions
		m.stagedUserMacroActions = nil
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Recorded macro",
		})
	} else {
		log.Printf("Started recording user macro\n")
		m.isRecordingUserMacro = true
		m.stagedUserMacroActions = nil
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleSuccess,
			Text:  "Started recording macro",
		})
	}
}

// AddToRecordingUserMacro adds an action to the currently recording user macro, if any.
func AddToRecordingUserMacro(s *EditorState, action MacroAction) {
	m := &s.macroState
	if m.isRecordingUserMacro {
		m.stagedUserMacroActions = append(m.stagedUserMacroActions, action)
	}
}

// ReplayRecordedUserMacro replays the recorded user-defined macro.
// If no macro has been recorded, this shows an error status msg.
func ReplayRecordedUserMacro(s *EditorState) {
	m := &s.macroState

	if m.isRecordingUserMacro {
		// Replaying a macro while recording a macro would cause unexpected results.
		// On the initial recording, the replay would refer to the previously-recorded macro,
		// but on subsequent replays it would refer to the newly-recorded macro.
		// Avoid this problem by disallowing replay while recording entirely.
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "Cannot replay a macro while recording a macro",
		})
		return
	}

	if len(m.userMacroActions) == 0 {
		SetStatusMsg(s, StatusMsg{
			Style: StatusMsgStyleError,
			Text:  "No macro has been recorded",
		})
		return
	}

	// Copy the actions into a new slice to ensure later recordings
	// do not change the behavior of the replay action.
	replayActions := make([]MacroAction, len(m.userMacroActions))
	copy(replayActions, m.userMacroActions)

	// Define a new action that replays the macro.
	// The action sets the isReplayingUserMacro flag to disable undo log checkpointing
	// when switching input modes -- this ensures that the next undo operation reverts
	// the entire macro.
	replay := func(s *EditorState) {
		log.Printf("Replaying actions from user macro...\n")
		s.macroState.isReplayingUserMacro = true
		for _, action := range m.userMacroActions {
			action(s)
		}
		log.Printf("Finished replaying actions from user macro\n")
		s.macroState.isReplayingUserMacro = false
	}

	// Replay the macro, then set the replay action as the new "last" action macro.
	// This lets the user easily repeat the macro using the "." command in normal mode.
	replay(s)
	m.lastActions = []MacroAction{replay}

	SetStatusMsg(s, StatusMsg{
		Style: StatusMsgStyleSuccess,
		Text:  "Replayed macro",
	})
}
