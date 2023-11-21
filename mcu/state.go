package mcu

import (
	"fmt"
	"log"
	"time"

	"github.com/normen/obs-mcu/gomcu"
	"gitlab.com/gomidi/midi/v2"
)

type McuState struct {
	// TODO: combine
	FaderLevels         []int16
	FaderLevelsBuffered []float64
	FaderTouch          []bool
	FaderTouchTimeout   []time.Time
	LedStates           map[byte]bool
	VPotLedStates       map[byte]byte
	Text                string
	Assign              []rune
	Debug               bool
}

func NewMcuState() *McuState {
	//time.Now().Add
	state := McuState{}
	state.Text = "                                                                                                                "
	state.Assign = []rune{' ', ' '}
	state.FaderLevels = append(state.FaderLevels, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	state.FaderLevelsBuffered = append(state.FaderLevelsBuffered, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	state.FaderTouch = []bool{false, false, false, false, false, false, false, false, false}
	//state.FaderTouchTimeout = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}
	now := time.Now()
	state.FaderTouchTimeout = []time.Time{now, now, now, now, now, now, now, now, now}
	state.LedStates = make(map[byte]bool)
	state.VPotLedStates = make(map[byte]byte)
	return &state
}

func (m *McuState) Update() {
	for i, level := range m.FaderLevelsBuffered {
		if m.FaderTouch[i] {
			now := time.Now()
			timeout := m.FaderTouchTimeout[i]
			since := now.Sub(timeout)
			if since.Milliseconds() > 300 {
				m.FaderTouch[i] = false
				// sends if not already same
				//log.Printf("Fader #%v is released, sending", i)
				m.SetFaderLevel(byte(i), level)
			}
		}
	}
}

func (m *McuState) SetFaderTouched(fader byte) {
	state.FaderTouch[fader] = true
	state.FaderTouchTimeout[fader] = time.Now()
}

func (m *McuState) SetFaderLevel(fader byte, level float64) {
	m.FaderLevelsBuffered[fader] = level
	newLevel := FaderFloatToInt(level)
	if m.FaderLevels[fader] != newLevel {
		m.FaderLevels[fader] = newLevel
		channel := gomcu.Channel(fader)
		if !m.FaderTouch[fader] {
			x := []midi.Message{gomcu.SetFaderPos(channel, uint16(newLevel))}
			SendMidi(x)
			if m.Debug {
				log.Print(x)
			}
		} else {
			//log.Printf("Fader #%v is touched, not sending", fader)
		}
	}
}

func (m *McuState) SetMonitorState(fader byte, state string) {
	// OBS_MONITORING_TYPE_NONE
	// OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT
	// OBS_MONITORING_TYPE_MONITOR_ONLY
	num := byte(gomcu.Rec1) + fader
	num2 := byte(gomcu.Solo1) + fader
	switch state {
	case "OBS_MONITORING_TYPE_NONE":
		m.SendLed(num, false)
		m.SendLed(num2, false)
	case "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT":
		m.SendLed(num, false)
		m.SendLed(num2, true)
	case "OBS_MONITORING_TYPE_MONITOR_ONLY":
		m.SendLed(num, true)
		m.SendLed(num2, false)
	}
}

func (m *McuState) SetMuteState(fader byte, state bool) {
	num := byte(gomcu.Mute1) + fader
	m.SendLed(num, state)
}

func (m *McuState) SendLed(num byte, state bool) {
	if m.LedStates[num] != state {
		//log.Printf("Sending led %v, %t", num, state)
		m.LedStates[num] = state
		var mstate gomcu.State
		if state {
			mstate = gomcu.StateOn
		} else {
			mstate = gomcu.StateOff
		}
		x := []midi.Message{gomcu.SetLED(gomcu.Switch(num), mstate)}
		SendMidi(x)
		if m.Debug {
			log.Print(x)
		}
	}
}

func (m *McuState) SetAssignText(text []rune) {
	if m.Assign[0] != text[0] || m.Assign[1] != text[1] {
		x := []midi.Message{gomcu.SetDigit(gomcu.AssignLeft, gomcu.Char(text[0])), gomcu.SetDigit(gomcu.AssignRight, gomcu.Char(text[1]))}
		SendMidi(x)
		m.Assign = text
		if m.Debug {
			log.Print(x)
		}
	}
}

func (m *McuState) SetChannelText(fader byte, text string, lower bool) {
	idx := int(fader * 7)
	if lower {
		idx += 56
	}
	text = ShortenText(text)
	if m.Text[idx:idx+6] != text {
		m.Text = fmt.Sprintf("%s%s%s", m.Text[0:idx], text, m.Text[idx+6:])
		x := []midi.Message{gomcu.SetLCD(idx, text)}
		SendMidi(x)
		if m.Debug {
			log.Print(x)
		}
	}
}
