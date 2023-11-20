package obs

import (
	"log"
	"sort"
	"time"

	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/normen/obs-mcu/msg"
)

type Channel struct {
	Name        string
	Visible     bool
	Muted       bool
	Pan         float64
	Volume      float64
	MonitorType string
	Tracks      []bool
}

type ChangeSet struct {
	ChangedChannels []Channel
}

type ChannelList struct {
	inputs       map[string]*Channel
	FirstChannel int
	syncRetry    *time.Timer
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
		return channels[l.FirstChannel:]
	} else {
		return []Channel{}
	}
}

func (l *ChannelList) GetVisibleName(index byte) string {
	visible := l.GetVisible()
	myIndex := int(index) + l.FirstChannel
	if len(visible) > myIndex {
		return visible[index].Name
	}
	return ""
}

func (l *ChannelList) GetVisibleNumber(name string) int {
	visible := l.GetVisible()
	for idx, ch := range visible {
		if ch.Name == name {
			myIdx := idx - l.FirstChannel
			if myIdx >= 0 {
				return myIdx
			} else {
				return -1
			}
		}
	}
	return -1
}

func (l *ChannelList) AddChannel(name string) {
	if _, ok := l.inputs[name]; !ok {
		c := Channel{Name: name, Volume: 0, Pan: 0, Muted: true}
		l.inputs[name] = &c
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
			l.getBaseInfos(name)
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
		}
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

func (l *ChannelList) SetTracks(name string, tracksEnabled map[int]bool) {
	if channel, ok := l.inputs[name]; ok {
		for i, enabled := range channel.Tracks {
			if enabled != tracksEnabled[i] {
				channel.Tracks[i] = tracksEnabled[i]
				break
			}
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
	l.syncRetry = time.AfterFunc(100*time.Millisecond, func() { sync <- 0 })
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
	}
	asgn := []rune{'0' + rune((l.FirstChannel+1)/10%10), '0' + rune((l.FirstChannel+1)%10)}
	fromObs <- msg.AssignLEDMessage{
		Characters: asgn,
	}
}

// TODO: other way to get active ones initially?
func (l *ChannelList) UpdateVisible() {
	//TODO: calling this here to set them enabled, make this better!
	resp, err := client.Scenes.GetCurrentProgramScene()
	if err == nil {
		list, err2 := client.SceneItems.GetSceneItemList(&sceneitems.GetSceneItemListParams{SceneName: resp.CurrentProgramSceneName})
		if err2 == nil {
			for _, item := range list.SceneItems {
				if item.SceneItemEnabled {
					l.setVisible(item.SourceName, true)
				}
				if item.SourceType == "OBS_SOURCE_TYPE_SCENE" {
					sublist, err3 := client.SceneItems.GetGroupSceneItemList(&sceneitems.GetGroupSceneItemListParams{SceneName: item.SourceName})
					if err3 == nil {
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
	//TODO here?
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
			//l.SetTracks(inputName, tracks.InputAudioTracks)
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
			//l.SetTracks(inputName, tracks.InputAudioTracks)
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

// TODO: reduce calls to this
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
	//TODO:TRACKS
	//inputList.SetTracks(inputName, tracks.InputAudioTracks)
	//log.Println(tracks.InputAudioTracks)
	// TODO: get the rest of the info
}
