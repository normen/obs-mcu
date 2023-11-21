package main

/**
Compile Linux:
sudo apt install clang libasound2-dev
**/

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/mcu"
	"github.com/normen/obs-mcu/obs"
)

var VERSION string = "v0.3.7"
var interrupt chan os.Signal

// TODO: config file command line option
func main() {
	var showMidi bool
	var showHelp bool
	flag.BoolVar(&showMidi, "l", false, "List all installed MIDI devices")
	flag.BoolVar(&showHelp, "h", false, "Show Help")
	flag.Parse()
	log.Printf("OBS-MCU %v", VERSION)
	if showHelp {
		fmt.Println("Usage: obs-mcu [options]")
		flag.PrintDefaults()
	} else if showMidi {
		ShowMidiPorts()
		config.InitConfig()
	} else {
		config.InitConfig()
		interrupt = make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		fromMcu := make(chan interface{}, 100)
		fromObs := make(chan interface{}, 100)
		obs.InitObs(fromMcu, fromObs)
		mcu.InitMcu(fromMcu, fromObs)
		<-interrupt
	}
}

func ShowMidiPorts() {
	inputs := mcu.GetMidiInputs()
	for _, v := range inputs {
		fmt.Printf("MIDI Input: %s\n", v)
	}
	outputs := mcu.GetMidiOutputs()
	for _, v := range outputs {
		fmt.Printf("MIDI Output: %s\n", v)
	}
}
