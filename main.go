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
	"os/signal"
	"strconv"
	"strings"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/mcu"
	"github.com/normen/obs-mcu/obs"
)

var VERSION string = "v0.4.1"
var interrupt chan os.Signal

// TODO: config file command line option
func main() {
	var showMidi, configureMidi, showHelp bool
	flag.BoolVar(&showMidi, "l", false, "List all installed MIDI devices")
	flag.BoolVar(&configureMidi, "c", false, "Configure and start")
	flag.BoolVar(&showHelp, "h", false, "Show Help")
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
	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	fromMcu := make(chan interface{}, 100)
	fromObs := make(chan interface{}, 100)
	obs.InitObs(fromMcu, fromObs)
	mcu.InitMcu(fromMcu, fromObs)
	<-interrupt
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
		fmt.Printf("MIDI Input #%v: %s\n", i+1, v)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Print("Enter input port number and press [enter]: ")
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
		fmt.Printf("MIDI Output #%v: %s\n", i+1, v)
	}
	fmt.Println()
	fmt.Print("Enter output port number and press [enter]: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	num, err = strconv.Atoi(text)
	if err != nil || num <= 0 || num > len(inputs) {
		fmt.Println("Please enter only valid numbers")
		return false
	}
	config.Config.Midi.PortOut = outputs[num-1]

	fmt.Println()
	fmt.Println()
	fmt.Println("*** CONFIGURING OBS connection ***")
	fmt.Println()
	fmt.Println("Please enter the OBS host name and websocket password, for (current) press [enter]")
	fmt.Printf("Enter host and port (%v): ", config.Config.General.ObsHost)
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text != "" {
		config.Config.General.ObsHost = text
	}
	fmt.Println()
	fmt.Printf("Enter password or press [enter] to keep current: ")
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
