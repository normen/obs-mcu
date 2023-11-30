package msg

// obs <- mackie
type UpdateRequest struct {
}

// obs <- mackie
type KeyMessage struct {
	HotkeyName string
}

// obs <- mackie
type BankMessage struct {
	ChangeAmount int
}

// obs <- mackie
type VPotChangeMessage struct {
	FaderNumber  byte
	ChangeAmount int
}

// obs <- mackie
type VPotButtonMessage struct {
	FaderNumber byte
}

// obs <-> mackie
type MuteMessage struct {
	FaderNumber byte
	Value       bool
}

// obs <-> mackie
type SelectMessage struct {
	FaderNumber byte
	Value       bool
}

// obs <-> mackie
type AssignMessage struct {
	Mode byte
}

// obs <-> mackie
type TrackEnableMessage struct {
	TrackNumber byte
	Value       bool
}

// obs <-> mackie
type FaderMessage struct {
	FaderNumber byte
	FaderValue  float64
}

// obs <-> mackie
type MonitorTypeMessage struct {
	FaderNumber byte
	MonitorType string
}

// obs -> mackie
type LedMessage struct {
	LedName  string
	LedState bool
}

// obs -> mackie
type ChannelTextMessage struct {
	FaderNumber byte
	Text        string
	Lower       bool
}

// obs -> mackie
type DisplayTextMessage struct {
	Text string
}

// obs -> mackie
type VPotLedMessage struct {
	FaderNumber byte
	LedState    byte
}

// obs -> mackie
type AssignLEDMessage struct {
	Characters []rune
}

// obs -> mackie
type MeterMessage struct {
	FaderNumber byte
	Value       float64
}

// internal mcu message
type RawFaderMessage struct {
	FaderNumber byte
	FaderValue  int16
}

// internal mcu message
type RawFaderTouchMessage struct {
	FaderNumber byte
	Pressed     bool
}
