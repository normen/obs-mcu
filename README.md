## OBS-MCU

Connect a Mackie Control (or compatible MCU) to OBS

#### What

This small application creates a bridge between OBS and a Mackie Control (or compatible) fader controller. It allows controlling the OBS audio channels through the hardware faders as well as executing OBS keyboard shortcuts via buttons on the control surface.

Its written in golang so the executable has pretty much no external dependencies and can be run as is on any system.

I created this for myself but it was in a well enough state (configurable etc) that I decided to release it, it comes with no warranties.

#### How

The application runs standalone alongside of OBS, it connects via MIDI to the MCU controller and via websockets to OBS. It then allows controlling the OBS audio channels as well as trigger any mapped buttons in OBS.

The fact that it runs as a standalone app even on a Raspberry Pi allows you to control your OBS from anywhere using your MCU, simply by running the app on a headless Pi and connecting it to your MCU and Wifi.

#### Controls

##### Fader Section

The `Faders` and the `Mute` buttons work as you'd expect, they basically mirror the audio mixer in OBS.

The `Solo` buttons set the monitor mode for the channel to "monitor and output".

The `Rec` buttons set the monitor mode for the channel to "monitor only".

If neither Rec nor Solo are lit the monitor mode is "off".

You can use the `Channel/Bank` buttons to see more channels in case you have more than 8 audio sources. The displays show the names of the channels, shortened to fit the MCU default length of 6 characters.

The rest of the fader section including the assign buttons are not mappable as they are kept free for future feature updates.

##### Buttons

Almost all other buttons except the assign and automation buttons are freely assignable through the config file (see below) but the standard mapping is this:

- `Play` - Start stream
- `Stop` - Stop stream
- `Rec` - Start recording

To map a button you have to find the internal OBS key name and assign it in the config file, prefixed with `KEY:`, like such:

```
[mcu_buttons]
play = KEY:OBSBasic.StartStreaming
```

#### Configuration

##### Basic Rundown

- Put `obs-mcu`(`.exe`) executable somewhere
- Connect the MCU
- Run `obs-mcu -l` to list all MIDI devices (this also creates the config file)
- Copy MIDI device name and enter in config file (see below)
- Enter OBS host and password in config file
- Run `obs-mcu`
- Run OBS and control your audio channels

##### Config file location

All configuration happens through a config file. The config file is created on the first start, its location is

- On Windows
  - `(HOME)\AppData\Local\obs-mcu`
- On MacOS
  - `(HOME)/Library/Application Support/obs-mcu`
- On Linux
  - `(HOME)/.config/obs-mcu`

You have to specify the OBS host, its password, and the MIDI in and out ports.

#### Command line options

- `-l` lists the names of all MIDI ports to help with the configuration

#### Caveats

Handling of MIDI device disconnects is currently not very graceful, you might have to restart the app when the MIDI ports change (or it crashes all by itself 🙂). Also you have to connect the MIDI device before starting the app. An update of the go-midi library to v2 should hopefully remedy this.

Theres afaict no way to get the "hidden" state of audio channels, so they will always display on the MCU even if they're hidden in OBS. As a workaround you can simply name these channels so that they appear all the way on the left and then press the "channel right" button until all channels you don't want to see are hidden.

#### TODO / Future

Heres a list of planned features I might get around to work on:

##### Features

- [ ] Allow loading different config files
- [ ] Select channels (for feature below)
- [ ] Allow selection of output tracks for selected channel
- [ ] Allow use of Vpots to control delay, show in display
- [ ] Allow assigning states in OBS to LEDs
- [ ] Show meters
- [ ] Video fade on master fader

##### Internal

- [ ] Update to go-midi v2 and better handle MIDI disconnects

#### Development

Building should be straightforward, on linux you need to have libasound2-dev available.

- Install golang
- Run `go build` or use the Makefile with `make`

##### Overview

Theres basically two runloops, one for the connection to the MCU and one for the connection to OBS. The latter is handling most of the logic while the MCU loop is basically just translating and trying to keep the amount of data being sent via MIDI low. They communicate through two channels via Message structs and keep draining each others messages while communicating with the MCU respectively OBS.

Having just two runloops instead of a heap of go routines makes it easy to track the control flow logic so theres no need for excessive locking and backckecking.

#### Thanks

Thanks to [chabad360](https://github.com/chabad360) for his [gomcu](https://github.com/chabad360/gomcu) code which is integrated in this app.
