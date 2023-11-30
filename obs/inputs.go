package obs

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/normen/obs-mcu/msg"
)

const (
	ModeDelay byte = iota
	Mode_2
	ModePan
	Mode_4
	Mode_5
	Mode_6
)

type Channel struct {
	Name        string
	Visible     bool
	Muted       bool
	Pan         float64
	Volume      float64
	MonitorType string
	DelayMS     float64
	Tracks      map[string]bool
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:   name,
		Tracks: make(map[string]bool),
	}
}

type ChannelList struct {
	inputs          map[string]*Channel
	FirstChannel    int
	AssignMode      byte
	SelectedChannel string
	syncRetry       *time.Timer
}

func NewChannelList() *ChannelList {
	return &ChannelList{
		inputs: make(map[string]*Channel),
	}
}

func (l *ChannelList) ChangeFaderBank(amount int) {
	l.FirstChannel = l.FirstChannel + amount
	if l.FirstChannel < 0 {
		l.FirstChannel = 0
	}
	l.sync()
}

// create alphabetically sorted list of visible channels
func (l *ChannelList) GetVisible() []Channel {
	var keys []string
	var channels []Channel
	for k := range l.inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		value, _ := l.inputs[k]
		if value.Visible {
			channels = append(channels, *value)
		}
	}
	if len(channels) > l.FirstChannel {
		vis := channels[l.FirstChannel:]
		if len(vis) > 8 {
			vis = vis[:8]
		}
		return vis
	} else {
		return []Channel{}
	}
}

func (l *ChannelList) GetVisibleName(index byte) string {
	visible := l.GetVisible()
	myIndex := int(index)
	if len(visible) > myIndex {
		return visible[index].Name
	}
	return ""
}

func (l *ChannelList) GetVisibleNumber(name string) int {
	visible := l.GetVisible()
	for idx, ch := range visible {
		if ch.Name == name {
			if idx >= 0 {
				return idx
			} else {
				return -1
			}
		}
	}
	return -1
}

func (l *ChannelList) AddChannel(name string) {
	if _, ok := l.inputs[name]; !ok {
		c := NewChannel(name)
		l.inputs[name] = c
		l.getBaseInfos(name)
	}
}

func (l *ChannelList) RemoveChannel(name string) {
	if _, ok := l.inputs[name]; ok {
		delete(l.inputs, name)
		l.sync()
	}
}

func (l *ChannelList) SetVisible(name string, visible bool) {
	if channel, ok := l.inputs[name]; ok {
		if channel.Visible != visible {
			channel.Visible = visible
			l.sync()
		}
	}
}

func (l *ChannelList) setVisible(name string, visible bool) {
	if channel, ok := l.inputs[name]; ok {
		if channel.Visible != visible {
			channel.Visible = visible
		}
	}
}

func (l *ChannelList) SetMuted(name string, muted bool) {
	if channel, ok := l.inputs[name]; ok {
		if channel.Muted != muted {
			channel.Muted = muted
			num := l.GetVisibleNumber(name)
			if num != -1 {
				fromObs <- msg.MuteMessage{
					FaderNumber: byte(num),
					Value:       muted,
				}
			}
		}
	}
}

func (l *ChannelList) SetPan(name string, pan float64) {
	if channel, ok := l.inputs[name]; ok {
		if channel.Pan != pan {
			channel.Pan = pan
			if l.AssignMode == ModePan {
				num := l.GetVisibleNumber(name)
				if num != -1 {
					fromObs <- msg.ChannelTextMessage{
						FaderNumber: byte(num),
						Lower:       true,
						Text:        fmt.Sprintf("%.2f", pan-0.5),
					}
					fromObs <- msg.VPotLedMessage{
						FaderNumber: byte(num),
						LedState:    byte(pan*11.0 + 1),
					}
				}
			}
		}
	}
}

func (l *ChannelList) GetPan(name string) float64 {
	if channel, ok := l.inputs[name]; ok {
		return channel.Pan
	}
	return 0
}

func (l *ChannelList) SetSelected(fader byte, selected bool) {
	name := l.GetVisibleName(fader)
	if name != "" {
		l.SelectedChannel = name
		fromObs <- msg.SelectMessage{
			FaderNumber: fader,
			Value:       true,
		}
		l.sync()
	}
}

func (l *ChannelList) SetMonitorType(name string, mon string) {
	if channel, ok := l.inputs[name]; ok {
		if channel.MonitorType != mon {
			channel.MonitorType = mon
			number := l.GetVisibleNumber(name)
			if number != -1 {
				fromObs <- msg.MonitorTypeMessage{
					FaderNumber: byte(number),
					MonitorType: mon,
				}
			}
		}
	}
}

func (l *ChannelList) GetMonitorType(name string) string {
	if channel, ok := l.inputs[name]; ok {
		return channel.MonitorType
	}
	return ""
}

func (l *ChannelList) SetDelayMS(name string, delay float64) {
	//delay = delay
	if channel, ok := l.inputs[name]; ok {
		if channel.DelayMS != delay {
			channel.DelayMS = delay
			if l.AssignMode == ModeDelay {
				num := l.GetVisibleNumber(name)
				if num != -1 {
					fromObs <- msg.ChannelTextMessage{
						FaderNumber: byte(num),
						Lower:       true,
						Text:        fmt.Sprintf("%.0fms", delay),
					}
					fromObs <- msg.VPotLedMessage{
						FaderNumber: byte(num),
						LedState:    0x00,
					}
				}
			}
		}
	}
}

func (l *ChannelList) GetDelayMS(name string) float64 {
	if channel, ok := l.inputs[name]; ok {
		return channel.DelayMS
	}
	return 0.0
}

func (l *ChannelList) SetVolume(name string, volume float64) {
	if channel, ok := l.inputs[name]; ok {
		if channel.Volume != volume {
			channel.Volume = volume
			number := l.GetVisibleNumber(name)
			if number != -1 {
				fromObs <- msg.FaderMessage{
					FaderNumber: byte(number),
					FaderValue:  volume,
				}
			}
		}
	}
}

func (l *ChannelList) SetTrack(idx byte, state bool) *Channel {
	if channel, ok := l.inputs[l.SelectedChannel]; ok {
		strIdx := fmt.Sprintf("%v", idx+1)
		if stateCur, ok := channel.Tracks[strIdx]; ok {
			channel.Tracks[strIdx] = !stateCur
			if channel.Name == l.SelectedChannel {
				fromObs <- msg.TrackEnableMessage{
					TrackNumber: byte(idx),
					Value:       !stateCur,
				}
			}
			return channel
		}
	}
	return nil
}

func (l *ChannelList) SetTracks(name string, tracksEnabled map[string]bool) {
	if channel, ok := l.inputs[name]; ok {
		channel.Tracks = tracksEnabled
		if name == l.SelectedChannel {
			for i, enabled := range channel.Tracks {
				idx, err := strconv.Atoi(i)
				if err == nil {
					idx--
					fromObs <- msg.TrackEnableMessage{
						TrackNumber: byte(idx),
						Value:       enabled,
					}
				} else {
					log.Println(err)
				}
			}
		}
	}
}

func (l *ChannelList) SetAssignMode(mode byte) {
	if mode == ModeDelay || mode == ModePan {
		if l.AssignMode != mode {
			l.AssignMode = mode
			fromObs <- msg.AssignMessage{
				Mode: mode,
			}
			l.sync()
		}
	}
}

func (l *ChannelList) Clear() {
	l.inputs = make(map[string]*Channel)
	l.sync()
}

func (l *ChannelList) SetAllInvisible() {
	for _, channel := range l.inputs {
		channel.Visible = false
	}
}

func (l *ChannelList) sync() {
	if l.syncRetry != nil {
		l.syncRetry.Stop()
		l.syncRetry = nil
	}
	// TODO: spaghetti (sync)
	l.syncRetry = time.AfterFunc(100*time.Millisecond, func() { sync <- l.SyncMcu })
}

func (l *ChannelList) SyncMcu() {
	var maxidx int = 0
	for i, input := range l.GetVisible() {
		fromObs <- msg.FaderMessage{
			FaderNumber: byte(i),
			FaderValue:  input.Volume,
		}
		fromObs <- msg.MuteMessage{
			FaderNumber: byte(i),
			Value:       input.Muted,
		}
		fromObs <- msg.MonitorTypeMessage{
			FaderNumber: byte(i),
			MonitorType: input.MonitorType,
		}
		fromObs <- msg.ChannelTextMessage{
			FaderNumber: byte(i),
			Text:        input.Name,
		}
		switch l.AssignMode {
		case ModeDelay:
			fromObs <- msg.ChannelTextMessage{
				FaderNumber: byte(i),
				Lower:       true,
				Text:        fmt.Sprintf("%.0fms", input.DelayMS),
			}
			fromObs <- msg.VPotLedMessage{
				FaderNumber: byte(i),
				LedState:    0x00,
			}
		case ModePan:
			fromObs <- msg.ChannelTextMessage{
				FaderNumber: byte(i),
				Lower:       true,
				Text:        fmt.Sprintf("%.2f", input.Pan-0.5),
			}
			fromObs <- msg.VPotLedMessage{
				FaderNumber: byte(i),
				LedState:    byte(input.Pan*11.0 + 1),
			}
		}
		maxidx = i + 1
	}
	for i := maxidx; i < 8; i++ {
		fromObs <- msg.FaderMessage{
			FaderNumber: byte(i),
			FaderValue:  0,
		}
		fromObs <- msg.MuteMessage{
			FaderNumber: byte(i),
			Value:       false,
		}
		fromObs <- msg.MonitorTypeMessage{
			FaderNumber: byte(i),
			MonitorType: "OBS_MONITORING_TYPE_NONE",
		}
		fromObs <- msg.ChannelTextMessage{
			FaderNumber: byte(i),
			Text:        "",
		}
		fromObs <- msg.ChannelTextMessage{
			FaderNumber: byte(i),
			Lower:       true,
			Text:        "",
		}
		fromObs <- msg.VPotLedMessage{
			FaderNumber: byte(i),
			LedState:    0x00,
		}
	}
	// assign display
	asgn := []rune{'0' + rune((l.FirstChannel+1)/10%10), '0' + rune((l.FirstChannel+1)%10)}
	fromObs <- msg.AssignLEDMessage{
		Characters: asgn,
	}
	// assign buttons
	fromObs <- msg.AssignMessage{
		Mode: l.AssignMode,
	}
	// select button
	selectNo := l.GetVisibleNumber(l.SelectedChannel)
	if selectNo != -1 {
		fromObs <- msg.SelectMessage{
			FaderNumber: byte(selectNo),
			Value:       true,
		}
	} else {
		fromObs <- msg.SelectMessage{
			FaderNumber: 0,
			Value:       false,
		}
	}
	// track enabled buttons
	if channel, ok := l.inputs[l.SelectedChannel]; ok {
		for i, enabled := range channel.Tracks {
			idx, err := strconv.Atoi(i)
			if err == nil {
				idx--
				fromObs <- msg.TrackEnableMessage{
					TrackNumber: byte(idx),
					Value:       enabled,
				}
			} else {
				log.Println(err)
			}
		}
	} else {
		for i := 0; i < 6; i++ {
			fromObs <- msg.TrackEnableMessage{
				TrackNumber: byte(i),
				Value:       false,
			}
		}
	}
	//TODO: spaghetti
	states.SendAll()
}

// TODO: other way to get active ones initially?
func (l *ChannelList) UpdateVisible() {
	resp, err := client.Scenes.GetCurrentProgramScene()
	if err == nil {
		list, err := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{SceneName: resp.CurrentProgramSceneName})
		if err == nil {
			for _, item := range list.SceneItems {
				if item.SceneItemEnabled {
					l.setVisible(item.SourceName, true)
				}
				if item.SourceType == "OBS_SOURCE_TYPE_SCENE" {
					sublist, err := client.SceneItems.GetGroupSceneItemList(&sceneitems.GetGroupSceneItemListParams{SceneName: item.SourceName})
					if err == nil {
						for _, subItem := range sublist.SceneItems {
							if subItem.SceneItemEnabled {
								l.setVisible(subItem.SourceName, true)
							}
						}
					} else {
						log.Print(err)
					}
				}
			}
		} else {
			log.Print(err)
		}
	} else {
		log.Print(err)
	}
	l.sync()
}

// adds an input and gets the basic info (mute state, volume etc)
func (l *ChannelList) AddInput(inputName string) {
	if len(inputName) == 0 {
		return
	}
	if _, ok := l.inputs[inputName]; !ok {
		tracks, _ := client.Inputs.GetInputAudioTracks(&inputs.GetInputAudioTracksParams{InputName: inputName})
		if tracks.InputAudioTracks != nil {
			l.AddChannel(inputName)
			l.sync()
		}
	}
}

func (l *ChannelList) addInput(inputName string) {
	if len(inputName) == 0 {
		return
	}
	if _, ok := l.inputs[inputName]; !ok {
		tracks, _ := client.Inputs.GetInputAudioTracks(&inputs.GetInputAudioTracksParams{InputName: inputName})
		if tracks.InputAudioTracks != nil {
			l.AddChannel(inputName)
		}
	}
}

// TODO: check for changes
func (l *ChannelList) UpdateSpecialInputs() error {
	resp, err := client.Inputs.GetSpecialInputs()
	if err == nil {
		l.addSpecialInput(resp.Desktop1)
		l.addSpecialInput(resp.Desktop2)
		l.addSpecialInput(resp.Mic1)
		l.addSpecialInput(resp.Mic2)
		l.addSpecialInput(resp.Mic3)
		l.addSpecialInput(resp.Mic4)
	} else {
		return err
	}
	return nil
}

func (l *ChannelList) addSpecialInput(inputName string) {
	l.addInput(inputName)
	l.setVisible(inputName, true)
}

func (l *ChannelList) getBaseInfos(inputName string) {
	volume, err := client.Inputs.GetInputVolume(&inputs.GetInputVolumeParams{InputName: inputName})
	if err == nil {
		l.SetVolume(inputName, volume.InputVolumeMul)
	} else {
		log.Print(err)
	}
	muted, err := client.Inputs.GetInputMute(&inputs.GetInputMuteParams{InputName: inputName})
	if err == nil {
		l.SetMuted(inputName, muted.InputMuted)
	} else {
		log.Print(err)
	}
	pan, err := client.Inputs.GetInputAudioBalance(&inputs.GetInputAudioBalanceParams{InputName: inputName})
	if err == nil {
		l.SetPan(inputName, pan.InputAudioBalance)
	} else {
		log.Print(err)
	}
	mon, err := client.Inputs.GetInputAudioMonitorType(&inputs.GetInputAudioMonitorTypeParams{InputName: inputName})
	if err == nil {
		l.SetMonitorType(inputName, mon.MonitorType)
	} else {
		log.Print(err)
	}
	sync, err := client.Inputs.GetInputAudioSyncOffset(&inputs.GetInputAudioSyncOffsetParams{InputName: inputName})
	if err == nil {
		l.SetDelayMS(inputName, sync.InputAudioSyncOffset)
	} else {
		log.Print(err)
	}
	tracks, err := client.Inputs.GetInputAudioTracks(&inputs.GetInputAudioTracksParams{InputName: inputName})
	if err == nil {
		l.SetTracks(inputName, map[string]bool(*tracks.InputAudioTracks))
	} else {
		log.Print(err)
	}
}
