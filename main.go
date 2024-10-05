package main

/**
Compile Linux:
sudo apt install clang libasound2-dev
**/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/getlantern/systray"
	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/icon"
	"github.com/normen/obs-mcu/mcu"
	"github.com/normen/obs-mcu/msg"
	"github.com/normen/obs-mcu/obs"
	"github.com/skratchdot/open-golang/open"
)

var VERSION string = "v0.7.14"
var waitGroup sync.WaitGroup

// TODO: config file command line option
func main() {
	var showMidi, configureMidi, showHelp bool
	flag.BoolVar(&showMidi, "l", false, "List all installed MIDI devices")
	flag.BoolVar(&configureMidi, "c", false, "Configure and start")
	flag.BoolVar(&showHelp, "h", false, "Show Help")
	flag.BoolVar(&obs.ExitWithObs, "x", false, "Exit when OBS exits")
	flag.BoolVar(&obs.ShowHotkeyNames, "k", false, "Show all of OBS hotkey names after connecting")
	flag.Parse()
	log.Printf("OBS-MCU %v", VERSION)
	if showHelp {
		fmt.Println("Usage: obs-mcu [options]")
		flag.PrintDefaults()
	} else if showMidi {
		ShowMidiPorts()
	} else if configureMidi {
		config.InitConfig()
		if UserConfigure() {
			startRunloops()
		}
	} else {
		config.InitConfig()
		if config.Config.Midi.PortIn == "" {
			if UserConfigure() {
				startRunloops()
			}
		} else {
			startRunloops()
		}
	}
}

func startRunloops() {
	fromMcu := make(chan interface{}, 100)
	fromObs := make(chan interface{}, 100)
	obs.InitObs(fromMcu, fromObs, &waitGroup)
	mcu.InitMcu(fromMcu, fromObs, &waitGroup)
	if isHeadless() {
		waitGroup.Wait()
	} else {
		go func() {
			waitGroup.Wait()
			log.Print("Quitting systray..")
			systray.Quit()
		}()
		systray.Run(onReady, onExit)
	}
}

func ShowMidiPorts() {
	inputs := mcu.GetMidiInputs()
	for i, v := range inputs {
		fmt.Printf("MIDI Input #%v: %s\n", i+1, v)
	}
	outputs := mcu.GetMidiOutputs()
	for i, v := range outputs {
		fmt.Printf("MIDI Output #%v: %s\n", i+1, v)
	}
}

func UserConfigure() bool {
	fmt.Println("*** CONFIGURING MIDI ***")
	fmt.Println("")
	inputs := mcu.GetMidiInputs()
	for i, v := range inputs {
		fmt.Printf("MIDI Input %v: %s\n", i+1, v)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Print("Enter MIDI input port number and press [enter]: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	num, err := strconv.Atoi(text)
	if err != nil || num <= 0 || num > len(inputs) {
		fmt.Println("Please enter only valid numbers")
		return false
	}
	config.Config.Midi.PortIn = inputs[num-1]

	fmt.Println()
	outputs := mcu.GetMidiOutputs()
	for i, v := range outputs {
		fmt.Printf("MIDI Output %v: %s\n", i+1, v)
	}
	fmt.Println()
	fmt.Print("Enter MIDI output port number and press [enter]: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	num, err = strconv.Atoi(text)
	if err != nil || num <= 0 || num > len(outputs) {
		fmt.Println("Please enter only valid numbers")
		return false
	}
	config.Config.Midi.PortOut = outputs[num-1]

	fmt.Println()
	fmt.Println()
	fmt.Println("*** CONFIGURING OBS connection ***")
	fmt.Println()
	fmt.Printf("Enter OBS host and port or press enter for (%v): ", config.Config.General.ObsHost)
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text != "" {
		config.Config.General.ObsHost = text
	}
	fmt.Println()
	fmt.Printf("Enter OBS password or press [enter] to keep current: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text != "" {
		config.Config.General.ObsPassword = text
	}
	fmt.Println()

	err = config.SaveConfig()
	if err != nil {
		log.Println(err)
	}
	return true
}

func onExit() {
}

func onReady() {
	fromUser := make(chan interface{}, 100)
	systray.SetTemplateIcon(icon.Data, icon.Data)
	//systray.SetTitle("obs-mcu")
	systray.SetTooltip("obs-mcu")
	mOpenConfig := systray.AddMenuItem("Edit Config", "Open config file")
	systray.AddSeparator()
	mMidiInputs := systray.AddMenuItem("MIDI Input", "Select MIDI input (restart to apply)")
	mMidiOutputs := systray.AddMenuItem("MIDI Output", "Select MIDI output (restart to apply)")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("Settings", "Other Settings")
	mShowMeters := mSettings.AddSubMenuItemCheckbox("Show Meters", "Show meters on MCU (restart to apply)", config.Config.McuFaders.ShowMeters)
	mSimulateTouch := mSettings.AddSubMenuItemCheckbox("Simulate Touch", "Simulate touch on MCU for surfaces with no touch support (restart to apply)", config.Config.McuFaders.SimulateTouch)
	inputs := mcu.GetMidiInputs()
	inputItems := make([]*systray.MenuItem, len(inputs))
	for i, v := range inputs {
		selected := config.Config.Midi.PortIn == v
		item := mMidiInputs.AddSubMenuItemCheckbox(v, "", selected)
		inputItems[i] = item
		val := v
		go func() {
			for {
				<-item.ClickedCh
				fromUser <- msg.MidiInputSetting{PortName: val}
				for _, v := range inputItems {
					v.Uncheck()
				}
				item.Check()
			}
		}()
	}
	outputs := mcu.GetMidiOutputs()
	outputItems := make([]*systray.MenuItem, len(outputs))
	for i, v := range outputs {
		selected := config.Config.Midi.PortOut == v
		item := mMidiOutputs.AddSubMenuItemCheckbox(v, "", selected)
		outputItems[i] = item
		val := v
		go func() {
			for {
				<-item.ClickedCh
				fromUser <- msg.MidiOutputSetting{PortName: val}
				for _, v := range outputItems {
					v.Uncheck()
				}
				item.Check()
			}
		}()
	}
	systray.AddSeparator()
	mQuitOrig := systray.AddMenuItem("Quit", "Quit obs-mcu")
	go func() {
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				systray.Quit()
			case <-mOpenConfig.ClickedCh:
				open.Run(config.GetConfigFilePath())
			case <-mShowMeters.ClickedCh:
				config.Config.McuFaders.ShowMeters = !config.Config.McuFaders.ShowMeters
				config.SaveConfig()
				if config.Config.McuFaders.ShowMeters {
					mShowMeters.Check()
				} else {
					mShowMeters.Uncheck()
				}
			case <-mSimulateTouch.ClickedCh:
				config.Config.McuFaders.SimulateTouch = !config.Config.McuFaders.SimulateTouch
				config.SaveConfig()
				if config.Config.McuFaders.SimulateTouch {
					mSimulateTouch.Check()
				} else {
					mSimulateTouch.Uncheck()
				}
			case message := <-fromUser:
				switch msg := message.(type) {
				case msg.MidiInputSetting:
					config.Config.Midi.PortIn = msg.PortName
					config.SaveConfig()
				case msg.MidiOutputSetting:
					config.Config.Midi.PortOut = msg.PortName
					config.SaveConfig()
				}
			}
		}
	}()
}

// check if we run headless
func isHeadless() bool {
	_, display := os.LookupEnv("DISPLAY")
	return runtime.GOOS != "windows" && runtime.GOOS != "darwin" && !display
}
