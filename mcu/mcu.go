package mcu

/*
TODO:
- update MIDI stack to v2
*/

import (
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/gomcu"
	"github.com/normen/obs-mcu/msg"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

var state *McuState
var drv *driver.Driver
var midiInput midi.In
var midiOutput midi.Out

var midiWriter *writer.Writer
var midiReader *reader.Reader
var connectRetry *time.Timer
var fromObs chan interface{}
var fromMcu chan interface{}
var internalMcu chan interface{}
var obsOutputChannel chan interface{}
var interrupt chan os.Signal
var connection chan int
var connected bool

// get a list of midi outputs
func GetMidiOutputs() []string {
	var names []string
	if drv, err := driver.New(); err == nil {
		defer drv.Close()
		if outs, erra := drv.Ins(); erra == nil {
			for _, output := range outs {
				names = append(names, output.String())
			}
		}
	}
	return names
}

// get a list of midi inputs
func GetMidiInputs() []string {
	var names []string
	if drv, err := driver.New(); err == nil {
		defer drv.Close()
		if ins, erra := drv.Ins(); erra == nil {
			for _, input := range ins {
				names = append(names, input.String())
			}
		}
	}
	return names
}

func InitMcu(fMcu chan interface{}, fObs chan interface{}) {
	fromMcu = fMcu
	fromObs = fObs
	InitInterp()
	internalMcu = make(chan interface{})
	interrupt = make(chan os.Signal, 1)
	connection = make(chan int, 1)
	signal.Notify(interrupt, os.Interrupt)
	state = NewMcuState()
	connection <- 0
	//connect()
	go runLoop()
}

func connect() {
	var err error
	disconnect()

	drv, err := driver.New()
	if err != nil {
		log.Print(err)
		retryConnect()
		return
	}

	ins, err := drv.Ins()
	if err != nil {
		log.Print(err)
		retryConnect()
		return
	}

	outs, err := drv.Outs()
	if err != nil {
		log.Print(err)
		retryConnect()
		return
	}
	for _, input := range ins {
		if input.String() == config.Config.Midi.PortIn {
			midiInput = input
		}
	}
	for _, output := range outs {
		if output.String() == config.Config.Midi.PortOut {
			midiOutput = output
		}
	}
	if midiInput == nil {
		log.Printf("Could not find MIDI Input '%s'", config.Config.Midi.PortIn)
		retryConnect()
		return
	}
	if midiOutput == nil {
		log.Printf("Could not find MIDI Output '%s'", config.Config.Midi.PortOut)
		retryConnect()
		return
	}

	err = midiInput.Open()
	if err != nil {
		log.Printf("Could not open MIDI Input '%s'", config.Config.Midi.PortOut)
		retryConnect()
	}
	err = midiOutput.Open()
	if err != nil {
		log.Printf("Could not open MIDI Output '%s'", config.Config.Midi.PortOut)
		retryConnect()
	}

	//TODO: check if closed
	midiWriter = writer.New(midiOutput)
	midiReader = reader.New(
		reader.NoLogger(),
		reader.NoteOn(noteon(midiWriter)),
		reader.Pitchbend(pitchbend(9)),
		reader.ControlChange(control(9)),
	)
	gomcu.Reset(midiWriter)

	go midiReader.ListenTo(midiInput)

	//m := []midi.Message{gomcu.SetDigit(gomcu.AssignLeft, 'H'), gomcu.SetDigit(gomcu.AssignRight, 'W'), gomcu.SetLCD(0, "Hello,"), gomcu.SetLCD(56, "World")}
	m := []midi.Message{}
	m = append(m, gomcu.SetTimeDisplay("OBS Studio")...)
	writer.WriteMessages(midiWriter, m)
	log.Print("MIDI Connected")
	connected = true
}

func disconnect() {
	connected = false
	//debug.PrintStack()
	if midiInput != nil {
		err := midiInput.Close()
		if err != nil {
			log.Print(err)
		}
		midiInput = nil
	}
	if midiOutput != nil {
		err := midiOutput.Close()
		if err != nil {
			log.Print(err)
		}
		midiOutput = nil
	}
	if drv != nil {
		err := drv.Close()
		if err != nil {
			log.Print(err)
		}
		drv = nil
	}
}

func retryConnect() {
	log.Print("Retry connection..")
	if connectRetry != nil {
		connectRetry.Stop()
	}
	connectRetry = time.AfterFunc(3*time.Second, func() { connection <- 0 })
}

func getCommand(k uint8) string {
	if len(gomcu.Names) > int(k) {
		fieldName := gomcu.Names[k]
		s := reflect.ValueOf(config.Config.McuButtons).Elem()
		fieldVal := s.FieldByName(fieldName)
		if fieldVal.IsValid() && fieldVal.Kind() == reflect.String {
			command := fieldVal.String()
			log.Printf("Got button %s, Command: %s", fieldName, command)
			return command
		}
	}
	return ""
}

func noteon(wr *writer.Writer) func(p *reader.Position, c, k, v uint8) {
	return func(p *reader.Position, c, k, v uint8) {
		//log.Printf("Channel %v, Note %v, Value %v", c, k, v)
		if gomcu.Switch(k) >= gomcu.BankL && gomcu.Switch(k) <= gomcu.ChannelR {
			var amount int
			switch gomcu.Switch(k) {
			case gomcu.BankL:
				amount = -8
			case gomcu.BankR:
				amount = 8
			case gomcu.ChannelL:
				amount = -1
			case gomcu.ChannelR:
				amount = 1
			}
			fromMcu <- msg.BankMessage{
				ChangeAmount: amount,
			}
		} else if gomcu.Switch(k) >= gomcu.Fader1 && gomcu.Switch(k) <= gomcu.Fader8 {
			// fader touch - handle locally
			internalMcu <- msg.RawFaderTouchMessage{
				FaderNumber: k - 104,
				Pressed:     v == 127,
			}
		} else if gomcu.Switch(k) >= gomcu.Mute1 && gomcu.Switch(k) <= gomcu.Mute8 {
			fromMcu <- msg.MuteMessage{
				FaderNumber: k - 0x10,
			}
		} else if gomcu.Switch(k) >= gomcu.Rec1 && gomcu.Switch(k) <= gomcu.Rec8 {
			fromMcu <- msg.MonitorTypeMessage{
				FaderNumber: k,
				MonitorType: "OBS_MONITORING_TYPE_MONITOR_ONLY",
			}
		} else if gomcu.Switch(k) >= gomcu.Solo1 && gomcu.Switch(k) <= gomcu.Solo8 {
			fromMcu <- msg.MonitorTypeMessage{
				FaderNumber: k - 0x08,
				MonitorType: "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT",
			}
		} else if len(gomcu.Names) > int(k) {
			command := getCommand(k)
			if len(command) > 0 {
				cmdType, cmdString, found := strings.Cut(command, ":")
				if found {
					switch cmdType {
					case "KEY":
						//send obs key
						fromMcu <- msg.KeyMessage{
							HotkeyName: cmdString,
						}
						//log.Printf("Got key press for %s", cmdString)
					}
				}
			}
		}
	}
}

func control(input int) func(p *reader.Position, c, k, value uint8) {
	return func(p *reader.Position, c, k, value uint8) {
		if gomcu.Switch(k) >= 0x10 && gomcu.Switch(k) <= 0x17 {
			amount := 0
			if value < 65 {
				amount = int(value)
			} else {
				amount = -1 * (int(value) - 64)
			}
			fromMcu <- msg.VPotChangeMessage{
				FaderNumber:  k - 0x10,
				ChangeAmount: amount,
			}
		}
	}
}

func pitchbend(input int) func(p *reader.Position, channel uint8, value int16) {
	return func(p *reader.Position, channel uint8, value int16) {
		//log.Printf("Value for fader #%d: %f", channel, value)
		internalMcu <- msg.RawFaderMessage{
			FaderNumber: channel,
			FaderValue:  value,
		}
		val := IntToFaderFloat(value)
		fromMcu <- msg.FaderMessage{
			FaderNumber: channel,
			FaderValue:  val,
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func checkMidiConnection() bool {
	if midiWriter == nil {
		return false
	}
	if midiInput != nil {
		if !midiInput.IsOpen() {
			retryConnect()
			return false
		}
	}
	return true
}

// only writes messages, reader is already looping
func runLoop() {
	timer := time.NewTicker(300 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			state.Update()
		case state := <-connection:
			if state == 0 {
				connect()
			}
		case <-interrupt:
			log.Print("Ending MCU runloop")
			//disconnect()
			return
		case message := <-fromObs:
			if !checkMidiConnection() {
				continue
			}
			switch e := message.(type) {
			case msg.FaderMessage:
				state.SetFaderLevel(e.FaderNumber, e.FaderValue)
			case msg.MuteMessage:
				state.SetMuteState(e.FaderNumber, e.Value)
			case msg.ChannelTextMessage:
				state.SetChannelText(e.FaderNumber, e.Text, e.Lower)
			case msg.AssignLEDMessage:
				state.SetAssignText(e.Characters)
			case msg.MonitorTypeMessage:
				state.SetMonitorState(e.FaderNumber, e.MonitorType)
			}
		case message := <-internalMcu:
			switch e := message.(type) {
			case msg.RawFaderMessage:
				state.SetFaderTouched(e.FaderNumber)
			case msg.RawFaderTouchMessage:
				state.SetFaderTouched(e.FaderNumber)
			}
		}
	}
}
