package obs

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/gorilla/websocket"

	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/events/subscriptions"
	"github.com/andreykaipov/goobs/api/requests/general"
	"github.com/andreykaipov/goobs/api/requests/inputs"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/msg"
)

var client *goobs.Client
var interrupt chan os.Signal
var connection chan int
var connected bool

//var connectCheck chan time.
var connectRetry *time.Timer
var channels *ChannelList
var fromMcu chan interface{}
var fromObs chan interface{}
var clientInputChannel chan interface{}

func InitObs(in chan interface{}, out chan interface{}) {
	fromMcu = in
	fromObs = out
	channels = NewChannelList()
	interrupt = make(chan os.Signal, 1)
	connection = make(chan int, 1)
	signal.Notify(interrupt, os.Interrupt)
	go runLoop()
	// start connection by sending connection state "0"
	connection <- 0
}

func connect() error {
	//interrupt <- os.Interrupt
	if client != nil {
		client.Disconnect()
	}
	var err error = nil
	// TODO: this basically blocks - the mackie channel could overflow
	client, err = goobs.New(config.Config.General.ObsHost,
		goobs.WithPassword(config.Config.General.ObsPassword),
		//goobs.WithEventSubscriptions(subscriptions.All|subscriptions.InputVolumeMeters|subscriptions.InputActiveStateChanged))
		goobs.WithEventSubscriptions(subscriptions.All|subscriptions.InputActiveStateChanged))
	if err != nil {
		return err
	}

	// Careful: changin this can only happen here because the loop calls connect()
	clientInputChannel = client.IncomingEvents

	version, err := client.General.GetVersion()
	if err != nil {
		return err
	}
	log.Printf("OBS Studio version: %s\n", version.ObsVersion)
	log.Printf("Websocket server version: %s\n", version.ObsWebSocketVersion)

	resp, _ := client.Inputs.GetInputList()
	for _, v := range resp.Inputs {
		channels.AddInput(v.InputName)
	}
	channels.UpdateSpecialInputs()
	channels.UpdateVisible()
	connected = true
	return nil
}

func retryConnect() {
	log.Print("Retry connection..")
	if connectRetry != nil {
		connectRetry.Stop()
	}
	connectRetry = time.AfterFunc(3*time.Second, func() { connection <- 0 })
}

func showInputs() {
	inputs := channels.GetVisible()
	for i, input := range inputs {
		log.Printf("Audio %d: %s", i, input.Name)
	}
}

func disconnect() {
	connected = false
	channels.Clear()
	if client != nil {
		client.Disconnect()
	}
}

func processMcuMessage(message interface{}) {
	if !connected {
		return
	}
	switch e := message.(type) {
	case msg.FaderMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			var err error
			//TODO: workaround for bug in goobs, setting 0 via mul doesn't work
			if e.FaderValue == 0 {
				_, err = client.Inputs.SetInputVolume(&inputs.SetInputVolumeParams{
					InputName:     name,
					InputVolumeDb: -100,
				})
			} else {
				_, err = client.Inputs.SetInputVolume(&inputs.SetInputVolumeParams{
					InputName:      name,
					InputVolumeMul: e.FaderValue,
				})
			}
			if err != nil {
				log.Print(err)
				log.Printf("Fader Volume: %v", e.FaderValue)
			}
		}
	case msg.MonitorTypeMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			mon := channels.GetMonitorType(name)
			var err error
			switch e.MonitorType {
			// can't come from the MCU
			//case "OBS_MONITORING_TYPE_NONE":
			case "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT":
				if mon == "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT" {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: name, MonitorType: "OBS_MONITORING_TYPE_NONE"})
				} else {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: name, MonitorType: "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT"})
				}
			case "OBS_MONITORING_TYPE_MONITOR_ONLY":
				if mon == "OBS_MONITORING_TYPE_MONITOR_ONLY" {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: name, MonitorType: "OBS_MONITORING_TYPE_NONE"})
				} else {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: name, MonitorType: "OBS_MONITORING_TYPE_MONITOR_ONLY"})
				}
			}
			if err != nil {
				log.Print(err)
			}
		}

	case msg.MuteMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			_, err := client.Inputs.ToggleInputMute(&inputs.ToggleInputMuteParams{InputName: name})
			if err != nil {
				log.Print(err)
			}
		}
	case msg.KeyMessage:
		_, err := client.General.TriggerHotkeyByName(&general.TriggerHotkeyByNameParams{HotkeyName: e.HotkeyName})
		if err != nil {
			log.Print(err)
		}
	case msg.BankMessage:
		channels.ChangeFaderBank(e.ChangeAmount)
	}
}

func processObsMessage(event interface{}) {
	switch e := event.(type) {
	//TODO: special inputs changed
	case *events.InputActiveStateChanged:
		//log.Printf("%s's active is now %t", e.InputName, e.VideoActive)
		//TODO: cache and deliver together
		channels.SetVisible(e.InputName, e.VideoActive)
	case *events.InputMuteStateChanged:
		//log.Printf("%s's mute is now %t", e.InputName, e.InputMuted)
		channels.SetMuted(e.InputName, e.InputMuted)
	case *events.InputVolumeChanged:
		//log.Printf("%s's volume is now %f", e.InputName, e.InputVolumeMul)
		channels.SetVolume(e.InputName, e.InputVolumeMul)
	case *events.InputNameChanged:
		//log.Printf("%s's name is now %s", e.OldInputName, e.InputName)
		// TODO: cleaner
		channels.RemoveChannel(e.OldInputName)
		channels.AddInput(e.InputName)
	case *events.InputAudioMonitorTypeChanged:
		//log.Printf("%s's monitor type is now %s", e.InputName, e.MonitorType)
		channels.SetMonitorType(e.InputName, e.MonitorType)
		// TODO: rec/solo
	case *events.InputCreated:
		//log.Printf("%s's been created", e.InputName)
		channels.AddInput(e.InputName)
	case *events.InputRemoved:
		//log.Printf("%s's been removed", e.InputName)
		channels.RemoveChannel(e.InputName)
	case *events.CurrentProgramSceneChanged:
	//log.Printf("Program change")
	//channels.UpdateVisible()
	case *events.InputAudioTracksChanged:
		//log.Printf("%s's audio tracks changed: %v", e.InputName, e.InputAudioTracks)
	case *events.ExitStarted:
		log.Print("Gracefully shutting down")
		//disconnect()
		connected = false
		//TODO: this is the only way we reconnect
		// -> other ways to see if connection dropped?
		channels.Clear()
		retryConnect()
		log.Print("Bye")
	case *events.InputVolumeMeters:
		log.Print(e.Inputs)
	case *websocket.CloseError:
		log.Print("OBS exited")
	default:
		//log.Printf("Unhandled: %#v", e)
		//log.Print(e)
	}
}

func handle(err error) {
	if err != nil {
		log.Print(err)
		retryConnect()
	}
}

func runLoop() {
	for {
		select {
		case <-interrupt:
			disconnect()
			log.Print("Ending OBS runloop")
			return
		case state := <-connection:
			switch state {
			case 0:
				handle(connect())
			}
		case message := <-fromMcu:
			processMcuMessage(message)
		case msg := <-clientInputChannel:
			processObsMessage(msg)
		}
	}
}
