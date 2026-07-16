package ui

// KeyBinding represents a single keyboard action with its associated keys and description.
type KeyBinding struct {
	Keys        []string
	Description string
}

// Matches returns true if the key message matches any of the mapped keys.
func (kb KeyBinding) Matches(msgStr string) bool {
	for _, k := range kb.Keys {
		if msgStr == k {
			return true
		}
		if k == "space" && msgStr == " " {
			return true
		}
	}
	return false
}

// KeyMap groups all keyboard shortcuts supported by PomoGo.
type KeyMap struct {
	Start        KeyBinding
	PauseResume  KeyBinding
	Skip         KeyBinding
	Task         KeyBinding
	Project      KeyBinding
	DeepFocus    KeyBinding
	ToggleStats  KeyBinding
	CopyStats    KeyBinding
	Reset        KeyBinding
	CycleTheme   KeyBinding
	CycleLayout  KeyBinding
	SoundPicker  KeyBinding
	ToggleZen    KeyBinding
	CycleEffects KeyBinding
	CycleVerb    KeyBinding
	Back         KeyBinding
	Quit         KeyBinding
	Help         KeyBinding
}

// ShortHelp returns a list of bindings to show in the help menu.
func (k KeyMap) ShortHelp() []KeyBinding {
	return []KeyBinding{
		k.Start,
		k.PauseResume,
		k.Skip,
		k.Task,
		k.Project,
		k.DeepFocus,
		k.ToggleStats,
		k.CopyStats,
		k.CycleTheme,
		k.CycleLayout,
		k.SoundPicker,
		k.ToggleZen,
		k.CycleEffects,
		k.CycleVerb,
		k.Reset,
		k.Back,
		k.Quit,
		k.Help,
	}
}

// DefaultKeyMap holds the default key bindings.
var DefaultKeyMap = KeyMap{
	Start:        KeyBinding{Keys: []string{"s"}, Description: "Start the focus session"},
	PauseResume:  KeyBinding{Keys: []string{"space"}, Description: "Pause / resume"},
	Skip:         KeyBinding{Keys: []string{"n"}, Description: "Skip to next phase"},
	Task:         KeyBinding{Keys: []string{"t"}, Description: "Set current task"},
	Project:      KeyBinding{Keys: []string{"p"}, Description: "Set current project"},
	DeepFocus:    KeyBinding{Keys: []string{"d"}, Description: "Choose Deep Focus duration"},
	ToggleStats:  KeyBinding{Keys: []string{"tab"}, Description: "Toggle statistics view"},
	CopyStats:    KeyBinding{Keys: []string{"y"}, Description: "Copy stats to clipboard"},
	Reset:        KeyBinding{Keys: []string{"r"}, Description: "Reset and clear state"},
	CycleTheme:   KeyBinding{Keys: []string{"T"}, Description: "Cycle theme"},
	CycleLayout:  KeyBinding{Keys: []string{"L"}, Description: "Cycle layout"},
	SoundPicker:  KeyBinding{Keys: []string{"a"}, Description: "Choose sound profile"},
	ToggleZen:    KeyBinding{Keys: []string{"S"}, Description: "Toggle screenshot mode"},
	CycleEffects: KeyBinding{Keys: []string{"e"}, Description: "Cycle ambient effects"},
	CycleVerb:    KeyBinding{Keys: []string{"v"}, Description: "Cycle activity verb"},
	Back:         KeyBinding{Keys: []string{"esc"}, Description: "Back / close overlay"},
	Quit:         KeyBinding{Keys: []string{"q", "ctrl+c"}, Description: "Quit"},
	Help:         KeyBinding{Keys: []string{"?"}, Description: "Toggle help overlay"},
}
