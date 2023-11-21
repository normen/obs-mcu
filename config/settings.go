package config

import (
	"log"
	"os"

	"github.com/adrg/xdg"
	"gopkg.in/ini.v1"
)

var configFilePath string
var cfg *ini.File

type IniFile struct {
	*General
	*Midi
	*McuFaders
	*McuVpots
	*McuLeds
	*McuButtons
}

type General struct {
	ObsHost     string
	ObsPassword string
}

type Midi struct {
	PortIn  string
	PortOut string
}

type McuFaders struct {
	ShowMeters bool
	//Fader1      string
	//Fader2      string
	//Fader3      string
	//Fader4      string
	//Fader5      string
	//Fader6      string
	//Fader7      string
	//Fader8      string
	//FaderMaster string
}

type McuVpots struct {
	//Vpot1 string
	//Vpot2 string
	//Vpot3 string
	//Vpot4 string
	//Vpot5 string
	//Vpot6 string
	//Vpot7 string
	//Vpot8 string
}

type McuLeds struct {
	//Rec1             string
	//Rec2             string
	//Rec3             string
	//Rec4             string
	//Rec5             string
	//Rec6             string
	//Rec7             string
	//Rec8             string
	//Solo1            string
	//Solo2            string
	//Solo3            string
	//Solo4            string
	//Solo5            string
	//Solo6            string
	//Solo7            string
	//Solo8            string
	//Mute1            string
	//Mute2            string
	//Mute3            string
	//Mute4            string
	//Mute5            string
	//Mute6            string
	//Mute7            string
	//Mute8            string
	//Select1          string
	//Select2          string
	//Select3          string
	//Select4          string
	//Select5          string
	//Select6          string
	//Select7          string
	//Select8          string
	//AssignTrack      string
	//AssignSend       string
	//AssignPan        string
	//AssignPlugin     string
	//AssignEQ         string
	//AssignInstrument string
	//Read    string
	//Write   string
	//Trim    string
	//Touch   string
	//Latch   string
	//Group   string
	Save    string
	Undo    string
	Marker  string
	Nudge   string
	Cycle   string
	Drop    string
	Replace string
	Click   string
	Solo    string
	Rewind  string
	FastFwd string
	Stop    string
	Play    string
	Record  string
	Zoom    string
	Scrub   string
}

type McuButtons struct {
	//Rec1             string
	//Rec2             string
	//Rec3             string
	//Rec4             string
	//Rec5             string
	//Rec6             string
	//Rec7             string
	//Rec8             string
	//Solo1            string
	//Solo2            string
	//Solo3            string
	//Solo4            string
	//Solo5            string
	//Solo6            string
	//Solo7            string
	//Solo8            string
	//Mute1            string
	//Mute2            string
	//Mute3            string
	//Mute4            string
	//Mute5            string
	//Mute6            string
	//Mute7            string
	//Mute8            string
	//Select1          string
	//Select2          string
	//Select3          string
	//Select4          string
	//Select5          string
	//Select6          string
	//Select7          string
	//Select8          string
	//V1               string
	//V2               string
	//V3               string
	//V4               string
	//V5               string
	//V6               string
	//V7               string
	//V8               string
	//AssignTrack      string
	//AssignSend       string
	//AssignPan        string
	//AssignPlugin     string
	//AssignEQ         string
	//AssignInstrument string
	//BankL           string
	//BankR           string
	//ChannelL        string
	//ChannelR        string
	//Flip            string
	//GlobalView      string
	//NameValue       string
	//SMPTEBeats      string
	F1              string
	F2              string
	F3              string
	F4              string
	F5              string
	F6              string
	F7              string
	F8              string
	MIDITracks      string
	Inputs          string
	AudioTracks     string
	AudioInstrument string
	Aux             string
	Busses          string
	Outputs         string
	User            string
	Shift           string
	Option          string
	Control         string
	CMDAlt          string
	//Read            string
	//Write           string
	//Trim            string
	//Touch           string
	//Latch           string
	//Group           string
	Save    string
	Undo    string
	Cancel  string
	Enter   string
	Marker  string
	Nudge   string
	Cycle   string
	Drop    string
	Replace string
	Click   string
	Solo    string
	Rewind  string
	FastFwd string
	Stop    string
	Play    string
	Record  string
	Up      string
	Down    string
	Left    string
	Right   string
	Zoom    string
	Scrub   string
	UserA   string
	UserB   string
	//Fader1           string
	//Fader2           string
	//Fader3           string
	//Fader4           string
	//Fader5           string
	//Fader6           string
	//Fader7           string
	//Fader8           string
	//FaderMaster      string
}

var Config = IniFile{
	&General{
		ObsHost:     "localhost:4455",
		ObsPassword: "",
	},
	&Midi{
		PortIn:  "MCU Mackie Control Port 1",
		PortOut: "MCU Mackie Control Port 1",
	},
	&McuFaders{
		ShowMeters: false,
		//Fader1:      "",
		//Fader2:      "",
		//Fader3:      "",
		//Fader4:      "",
		//Fader5:      "",
		//Fader6:      "",
		//Fader7:      "",
		//Fader8:      "",
		//FaderMaster: "",
	},
	&McuVpots{
		//Vpot1: "",
		//Vpot2: "",
		//Vpot3: "",
		//Vpot4: "",
		//Vpot5: "",
		//Vpot6: "",
		//Vpot7: "",
		//Vpot8: "",
	},
	&McuLeds{
		//Rec1:             "",
		//Rec2:             "",
		//Rec3:             "",
		//Rec4:             "",
		//Rec5:             "",
		//Rec6:             "",
		//Rec7:             "",
		//Rec8:             "",
		//Solo1:            "",
		//Solo2:            "",
		//Solo3:            "",
		//Solo4:            "",
		//Solo5:            "",
		//Solo6:            "",
		//Solo7:            "",
		//Solo8:            "",
		//Mute1:            "",
		//Mute2:            "",
		//Mute3:            "",
		//Mute4:            "",
		//Mute5:            "",
		//Mute6:            "",
		//Mute7:            "",
		//Mute8:            "",
		//Select1:          "",
		//Select2:          "",
		//Select3:          "",
		//Select4:          "",
		//Select5:          "",
		//Select6:          "",
		//Select7:          "",
		//Select8:          "",
		//AssignTrack:      "",
		//AssignSend:       "",
		//AssignPan:        "",
		//AssignPlugin:     "",
		//AssignEQ:         "",
		//AssignInstrument: "",
		//Read:    "",
		//Write:   "",
		//Trim:    "",
		//Touch:   "",
		//Latch:   "",
		//Group:   "",
		Save:    "",
		Undo:    "",
		Marker:  "",
		Nudge:   "",
		Cycle:   "",
		Drop:    "",
		Replace: "",
		Click:   "",
		Solo:    "",
		Rewind:  "",
		FastFwd: "",
		Stop:    "",
		Play:    "STATE:StreamState",
		Record:  "STATE:RecordState",
		Zoom:    "",
		Scrub:   "",
	},
	&McuButtons{
		//Rec1:             "",
		//Rec2:             "",
		//Rec3:             "",
		//Rec4:             "",
		//Rec5:             "",
		//Rec6:             "",
		//Rec7:             "",
		//Rec8:             "",
		//Solo1:            "",
		//Solo2:            "",
		//Solo3:            "",
		//Solo4:            "",
		//Solo5:            "",
		//Solo6:            "",
		//Solo7:            "",
		//Solo8:            "",
		//Mute1:            "",
		//Mute2:            "",
		//Mute3:            "",
		//Mute4:            "",
		//Mute5:            "",
		//Mute6:            "",
		//Mute7:            "",
		//Mute8:            "",
		//Select1:          "",
		//Select2:          "",
		//Select3:          "",
		//Select4:          "",
		//Select5:          "",
		//Select6:          "",
		//Select7:          "",
		//Select8:          "",
		//V1:               "",
		//V2:               "",
		//V3:               "",
		//V4:               "",
		//V5:               "",
		//V6:               "",
		//V7:               "",
		//V8:               "",
		//AssignTrack:      "",
		//AssignSend:       "",
		//AssignPan:        "",
		//AssignPlugin:     "",
		//AssignEQ:         "",
		//AssignInstrument: "",
		//BankL:           "",
		//BankR:           "",
		//ChannelL:        "",
		//ChannelR:        "",
		//Flip:            "",
		//GlobalView:      "",
		//NameValue:       "",
		//SMPTEBeats:      "",
		F1:              "",
		F2:              "",
		F3:              "",
		F4:              "",
		F5:              "",
		F6:              "",
		F7:              "",
		F8:              "",
		MIDITracks:      "",
		Inputs:          "",
		AudioTracks:     "",
		AudioInstrument: "",
		Aux:             "",
		Busses:          "",
		Outputs:         "",
		User:            "",
		Shift:           "",
		Option:          "",
		Control:         "",
		CMDAlt:          "",
		//Read:            "",
		//Write:           "",
		//Trim:            "",
		//Touch:           "",
		//Latch:           "",
		//Group:           "",
		Save:    "",
		Undo:    "",
		Cancel:  "",
		Enter:   "",
		Marker:  "",
		Nudge:   "",
		Cycle:   "",
		Drop:    "",
		Replace: "",
		Click:   "",
		Solo:    "",
		Rewind:  "",
		FastFwd: "",
		Stop:    "KEY:OBSBasic.ForceStopStreaming",
		Play:    "KEY:OBSBasic.StartStreaming",
		Record:  "KEY:OBSBasic.StartRecording",
		Up:      "",
		Down:    "",
		Left:    "",
		Right:   "",
		Zoom:    "",
		Scrub:   "",
		UserA:   "",
		UserB:   "",
		//Fader1:           "",
		//Fader2:           "",
		//Fader3:           "",
		//Fader4:           "",
		//Fader5:           "",
		//Fader6:           "",
		//Fader7:           "",
		//Fader8:           "",
		//FaderMaster:      "",
	},
}

func InitConfig() {
	var err error
	if configFilePath, err = xdg.ConfigFile("obs-mcu/obs-mcu.config"); err == nil {
		// add any new values
		var cfg *ini.File
		if cfg, err = ini.Load(configFilePath); err == nil {
			cfg.NameMapper = ini.TitleUnderscore
			cfg.ValueMapper = os.ExpandEnv
			if section, err := cfg.GetSection("general"); err == nil {
				section.MapTo(&Config.General)
			}
			if section, err := cfg.GetSection("midi"); err == nil {
				section.MapTo(&Config.Midi)
			}
			if section, err := cfg.GetSection("mcu_faders"); err == nil {
				section.MapTo(&Config.McuFaders)
			}
			if section, err := cfg.GetSection("mcu_vpots"); err == nil {
				section.MapTo(&Config.McuVpots)
			}
			if section, err := cfg.GetSection("mcu_leds"); err == nil {
				section.MapTo(&Config.McuLeds)
			}
			if section, err := cfg.GetSection("mcu_buttons"); err == nil {
				section.MapTo(&Config.McuButtons)
			}
			//TODO: only save if changes
			newCfg := ini.Empty()
			if err = ini.ReflectFromWithMapper(newCfg, &Config, ini.TitleUnderscore); err == nil {
				err = newCfg.SaveTo(configFilePath)
			}
		} else {
			cfg = ini.Empty()
			cfg.NameMapper = ini.TitleUnderscore
			cfg.ValueMapper = os.ExpandEnv
			if err = ini.ReflectFromWithMapper(cfg, &Config, ini.TitleUnderscore); err == nil {
				err = cfg.SaveTo(configFilePath)
			}
		}
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}

func GetConfigFilePath() string {
	return configFilePath
}
