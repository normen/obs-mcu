package msg

type ControlMessage struct {
	Type  string
	Value string
}

type StateMessage struct {
	Type  string
	Value string
}

// sent by mackie to obs

type KeyMessage struct {
	HotkeyName string
}

// sent by obs to mackie

type RecMessage struct {
	FaderNumber byte
	Value       bool
}

type SoloMessage struct {
	FaderNumber byte
	Value       bool
}

type MuteMessage struct {
	FaderNumber byte
	Value       bool
}

type SelectMessage struct {
	FaderNumber byte
	Value       bool
}

type FaderMessage struct {
	FaderNumber byte
	FaderValue  float64
}

type MonitorTypeMessage struct {
	FaderNumber byte
	MonitorType string
}

type VPotMessage struct {
	VPotNumber byte
	VPotValue  float64
}

type LedMessage struct {
	LedNumber byte
	LedValue  bool
}

type TextMessage struct {
	Text  string
	Lower bool
	Start byte
	End   byte
}

type ChannelTextMessage struct {
	FaderNumber byte
	Text        string
	Lower       bool
}

type BankMessage struct {
	ChangeAmount int
}

type AssignLEDMessage struct {
	Characters []rune
}

// internal mcu messages
type RawFaderMessage struct {
	FaderNumber byte
	FaderValue  int16
}

type RawFaderTouchMessage struct {
	FaderNumber byte
	Pressed     bool
}
