package obs

import (
	"errors"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/gorilla/websocket"

	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/events/subscriptions"
	"github.com/andreykaipov/goobs/api/requests/general"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/typedefs"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/msg"
)

var waitGroup *sync.WaitGroup
var ExitWithObs bool
var ShowHotkeyNames bool

var client *goobs.Client
var interrupt chan os.Signal
var connection chan int
var synch chan func()
var connected bool

var connectRetry *time.Timer
var channels *ChannelList
var states *ObsStates
var fromMcu chan interface{}
var fromObs chan interface{}
var clientInputChannel chan interface{}

var (
	OBS_MONITORING_TYPE_NONE               string = "OBS_MONITORING_TYPE_NONE"
	OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT string = "OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT"
	OBS_MONITORING_TYPE_MONITOR_ONLY       string = "OBS_MONITORING_TYPE_MONITOR_ONLY"
)

// Starts the runloop that manages the connection to OBS
func InitObs(in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
	fromMcu = in
	fromObs = out
	waitGroup = wg
	channels = NewChannelList()
	states = NewObsStates()
	// add always on state
	states.SetState("AlwaysOn", true)
	interrupt = make(chan os.Signal, 1)
	connection = make(chan int, 1)
	synch = make(chan func(), 1)
	signal.Notify(interrupt, os.Interrupt)
	wg.Add(1)
	go runLoop()
	// start connection by sending connection state "0"
	connection <- 0
}

// Tries to connect to OBS, called by the runloop
func connect() error {
	if client != nil {
		client.Disconnect()
	}
	var err error = nil
	// TODO: this basically blocks - the mackie channel could overflow
	if config.Config.ShowMeters {
		client, err = goobs.New(config.Config.General.ObsHost,
			goobs.WithPassword(config.Config.General.ObsPassword),
			goobs.WithEventSubscriptions(subscriptions.All|subscriptions.InputVolumeMeters|subscriptions.InputActiveStateChanged))

	} else {
		client, err = goobs.New(config.Config.General.ObsHost,
			goobs.WithPassword(config.Config.General.ObsPassword),
			goobs.WithEventSubscriptions(subscriptions.All|subscriptions.InputActiveStateChanged))
	}
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
	scene, err := client.Scenes.GetCurrentProgramScene(&scenes.GetCurrentProgramSceneParams{})
	if err == nil {
		fromObs <- msg.DisplayTextMessage{
			Text: scene.CurrentProgramSceneName,
		}
	}
	if ShowHotkeyNames {
		hotkeys, err := client.General.GetHotkeyList(&general.GetHotkeyListParams{})
		if err == nil {
			for _, key := range hotkeys.Hotkeys {
				log.Printf("KEY:%v", key)
			}
		}
	}
	connected = true
	return nil
}

// Tries to reconnect to OBS, called by the runloop
func retryConnect() {
	log.Print("Retry OBS connection..")
	if connectRetry != nil {
		connectRetry.Stop()
	}
	connectRetry = time.AfterFunc(3*time.Second, func() { connection <- 0 })
}

// Shows the current inputs in the log (for debugging)
func showInputs() {
	inputs := channels.GetVisible()
	for i, input := range inputs {
		log.Printf("Audio %d: %s", i, input.Name)
	}
}

// Disconnects from OBS, called by the runloop
func disconnect() {
	connected = false
	channels.Clear()
	if client != nil {
		client.Disconnect()
	}
}

// Processes a message from the MCU,
// called by the runloop when a message is received
func processMcuMessage(message interface{}) {
	if !connected {
		return
	}
	switch e := message.(type) {
	case msg.FaderMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			var err error
			_, err = client.Inputs.SetInputVolume(&inputs.SetInputVolumeParams{
				InputName:      &name,
				InputVolumeMul: &e.FaderValue,
			})
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
			case OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT:
				if mon == OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: &name, MonitorType: &OBS_MONITORING_TYPE_NONE})
				} else {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: &name, MonitorType: &OBS_MONITORING_TYPE_MONITOR_AND_OUTPUT})
				}
			case OBS_MONITORING_TYPE_MONITOR_ONLY:
				if mon == OBS_MONITORING_TYPE_MONITOR_ONLY {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: &name, MonitorType: &OBS_MONITORING_TYPE_NONE})
				} else {
					_, err = client.Inputs.SetInputAudioMonitorType(&inputs.SetInputAudioMonitorTypeParams{InputName: &name, MonitorType: &OBS_MONITORING_TYPE_MONITOR_ONLY})
				}
			}
			if err != nil {
				log.Print(err)
			}
		}
	case msg.MuteMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			_, err := client.Inputs.ToggleInputMute(&inputs.ToggleInputMuteParams{InputName: &name})
			if err != nil {
				log.Print(err)
			}
		}
	case msg.KeyMessage:
		_, err := client.General.TriggerHotkeyByName(&general.TriggerHotkeyByNameParams{HotkeyName: &e.HotkeyName})
		if err != nil {
			log.Print(err)
		}
	case msg.BankMessage:
		channels.ChangeFaderBank(e.ChangeAmount)
	case msg.SelectMessage:
		channels.SetSelected(e.FaderNumber, e.Value)
	case msg.AssignMessage:
		channels.SetAssignMode(e.Mode)
	case msg.TrackEnableMessage:
		channel := channels.SetTrack(e.TrackNumber, e.Value)
		if channel != nil {
			_, err := client.Inputs.SetInputAudioTracks(&inputs.SetInputAudioTracksParams{InputName: &channel.Name, InputAudioTracks: (*typedefs.InputAudioTracks)(&channel.Tracks)})
			if err != nil {
				log.Print(err)
			}
		}
	case msg.UpdateRequest:
		channels.SyncMcu()
	case msg.VPotButtonMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		balhalf := 0.5
		minval := 0.0
		if name != "" {
			switch channels.AssignMode {
			case ModePan:
				_, err := client.Inputs.SetInputAudioBalance(&inputs.SetInputAudioBalanceParams{InputName: &name, InputAudioBalance: &balhalf})
				if err != nil {
					log.Print(err)
				}
			case ModeDelay:
				_, err := client.Inputs.SetInputAudioSyncOffset(&inputs.SetInputAudioSyncOffsetParams{InputName: &name, InputAudioSyncOffset: &minval})
				if err != nil {
					log.Print(err)
				}
			}
		}
	case msg.VPotChangeMessage:
		name := channels.GetVisibleName(e.FaderNumber)
		if name != "" {
			switch channels.AssignMode {
			case ModePan:
				newPan := channels.GetPan(name) + float64(e.ChangeAmount)/50.0
				newPan = math.Min(newPan, 1.0)
				newPan = math.Max(newPan, 0.0)
				_, err := client.Inputs.SetInputAudioBalance(&inputs.SetInputAudioBalanceParams{InputName: &name, InputAudioBalance: &newPan})
				if err != nil {
					log.Print(err)
				}
			case ModeDelay:
				newDelay := channels.GetDelayMS(name) + float64(e.ChangeAmount*10)
				if newDelay < 10 && newDelay > -10 {
					newDelay = 0
				}
				_, err := client.Inputs.SetInputAudioSyncOffset(&inputs.SetInputAudioSyncOffsetParams{InputName: &name, InputAudioSyncOffset: &newDelay})
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
}

// Processes a message from OBS,
// called by the runloop when a message is received
func processObsMessage(event interface{}) {
	switch e := event.(type) {
	//TODO: special inputs changed
	case *events.InputActiveStateChanged:
		channels.SetVisible(e.InputName, e.VideoActive)
	case *events.InputMuteStateChanged:
		channels.SetMuted(e.InputName, e.InputMuted)
	case *events.InputVolumeChanged:
		channels.SetVolume(e.InputName, e.InputVolumeMul)
	case *events.InputNameChanged:
		// TODO: cleaner
		channels.RemoveChannel(e.OldInputName)
		channels.AddInput(e.InputName)
	case *events.InputAudioMonitorTypeChanged:
		channels.SetMonitorType(e.InputName, e.MonitorType)
	case *events.InputCreated:
		channels.AddInput(e.InputName)
	case *events.InputRemoved:
		channels.RemoveChannel(e.InputName)
	case *events.CurrentProgramSceneChanged:
		fromObs <- msg.DisplayTextMessage{
			Text: e.SceneName,
		}
	case *events.InputAudioTracksChanged:
		channels.SetTracks(e.InputName, map[string]bool(*e.InputAudioTracks))
	case *events.InputAudioBalanceChanged:
		channels.SetPan(e.InputName, e.InputAudioBalance)
	case *events.InputAudioSyncOffsetChanged:
		channels.SetDelayMS(e.InputName, e.InputAudioSyncOffset)
	case *events.ExitStarted:
		log.Print("OBS is shutting down")
		doExit()
	case *events.InputVolumeMeters:
		for _, v := range e.Inputs {
			num := channels.GetVisibleNumber(v.Name)
			if num != -1 {
				chNum := len(v.Levels)
				if chNum > 0 {
					valNum := len(v.Levels[0])
					if valNum > 0 {
						level := 0.0
						for i := 0; i < chNum; i++ {
							// TODO: what is what here? Levels[0][2] seems to be peak,
							// Levels[0][1] seems to be RMS?
							level = level + v.Levels[i][valNum-1]
							//log.Printf("%v ch[%d]: %f", v.Name, i, v.Levels[i][0])
						}
						level = level / float64(chNum)
						dbVal := 20 * math.Log10(level)
						fromObs <- msg.MeterMessage{
							FaderNumber: byte(num),
							Value:       dbVal,
						}
					}
				}
			}
		}
	case *events.StreamStateChanged:
		states.SetState("StreamState", e.OutputActive)
	case *events.RecordStateChanged:
		states.SetState("RecordState", e.OutputActive)
	case error:
		uw := errors.Unwrap(e)
		switch uw.(type) {
		case *websocket.CloseError:
			log.Print("OBS connection closed")
			doExit()
		default:
			log.Print(uw)
		}
	default:
		//log.Printf("Unhandled: %#v", e)
		//log.Print(e)
	}
}

func doExit() {
	channels.Clear()
	if ExitWithObs {
		log.Print("Bye")
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			log.Print(err)
		}
		if runtime.GOOS == "windows" {
			p.Kill()
		} else {
			p.Signal(syscall.SIGINT)
		}
	} else {
		retryConnect()
	}
}

// Handles an error by logging it and retrying to connect
func handle(err error) {
	if err != nil {
		log.Print(err)
		retryConnect()
	}
}

// The runloop that manages the connection to OBS
func runLoop() {
	for {
		select {
		case <-interrupt:
			disconnect()
			log.Print("Ending OBS runloop")
			waitGroup.Done()
			return
		case function := <-synch:
			function()
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
