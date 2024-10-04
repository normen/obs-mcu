## OBS-MCU

Connect a Mackie Control (or compatible MCU) to OBS

### What

This small application creates a bridge between OBS and a Mackie Control (or compatible) fader controller. It allows controlling the OBS audio channels through the hardware faders as well as executing OBS keyboard shortcuts via buttons on the control surface.

Most features of the OBS audio mixer are supported, including the ability to control the audio sync offset, the output tracks, the monitoring state and balance of the audio channels. It also displays the metering of the audio channels on the MCU.

### How

The application runs standalone alongside of OBS, it connects via MIDI to the MCU controller and via websockets to OBS. It then allows controlling the OBS audio channels as well as trigger any mapped buttons in OBS.

The fact that it runs as a standalone app even on a Raspberry Pi allows you to control your OBS from anywhere using your MCU, simply by running the app on a headless Pi and connecting it to your MCU and Wifi.

Its written in golang so the executable has pretty much no external dependencies and can be run as is on any system.

### Controls

#### Fader Section

The `Faders` and the `Mute` buttons work as you'd expect, they basically mirror the audio mixer in OBS.

The `VPots` allow you to change the audio sync offset for the audio channel in 10ms increments, the LCD shows the current offset. Pressing the `Pan` button in the assign section switches the VPots to control the balance of the audio channel. Pressing the VPot button resets the offset or balance to 0.

The `Solo` buttons set the monitor mode for the channel to "monitor and output".

The `Rec` buttons set the monitor mode for the channel to "monitor only".

If neither Rec nor Solo are lit the monitor mode is "off".

The `Select` buttons allow you to select a channed to display its enabled tracks in the automation section.

The automation section (buttons `Read` -> `Group`) allows enabling the single output stream tracks for an audio source.

You can use the `Channel/Bank` buttons to see more channels in case you have more than 8 audio sources. The displays show the names of the channels, shortened to fit the MCU default length of 6 characters.

The rest of the fader section including the assign buttons are not mappable as they are kept free for future feature updates.

#### Buttons

Almost all other buttons except the assign and automation buttons are freely assignable to any OBS keyboard shortcut through the config file (see below).

The standard mapping is as follows:

- `Play` - Start stream
- `Stop` - Stop stream
- `Rec` - Start recording

### Installation

On Windows and Linux just put the binary somewhere and run it, for MacOS theres a homebrew tap to install the latest version:

`brew install normen/tap/obs-mcu`

which can then be run from the command line using `obs-mcu`

### Configuration

#### Basic Rundown

- Enable websockets in OBS
- Put `obs-mcu`(`.exe`) executable somewhere
- Connect the MCU
- Run `obs-mcu`
- On the first start you will be able to configure the MIDI devices and OBS connection
  - Enter the number of your MIDI in and out device from the list
  - Enter the OBS host name or keep the default for local operation
  - Enter the OBS websocket password or keep empty for no password
- Start OBS and control your audio channels

#### Config file

All configuration happens through a config file. The config file is created on the first start, its location is

- On Windows
  - `(HOME)\AppData\Local\obs-mcu`
- On MacOS
  - `(HOME)/Library/Application Support/obs-mcu`
- On Linux
  - `(HOME)/.config/obs-mcu`

You have to specify the OBS host, its password, and the MIDI in and out ports.

##### Buttons

To map a button you have to find the internal OBS key name and assign it in the config file, prefixed with `KEY:`, like so:

```
[mcu_buttons]
play = KEY:OBSBasic.StartStreaming
```

##### LEDs

Some buttons have LEDs which can be assigned with states in OBS, the supported states are (for now):

- `StreamState`, when OBS is streaming
- `RecordState`, when OBS is recording
- `AlwaysOn`, to always light that LED

They have to be prefixed with `STATE:`, like so:

```
[mcu_leds]
play = STATE:StreamState
```

##### Fader Options

Set these options under `mcu_faders` to `true` to enable the respective feature:

- `show_meters` - Show the audio meters on the MCU (might be slow on slower systems)
- `simulate_touch` - Simulate a touch on the MCU fader when the fader is moved (for surfaces with no touch detection)

#### Command line options

- `-c` configure the basic MIDI and OBS connection settings
- `-l` lists the names of all MIDI ports
- `-k` lists the names of all OBS keyboard shortcuts (after connecting)
- `-x` exits obs-mcu when OBS exits 

#### Auto-Start with OBS

You can run obs-mcu automatically every time that OBS starts by using the OBS plugin [Advanced Scene Switcher](https://github.com/WarmUpTill/SceneSwitcher) in combination with the `-x` command line option. _Make sure you configure obs-mcu before you do this because you won't see the obs-mcu window when starting like this._

1. Install Advanced Scene Switcher
2. Add a macro to run obs-mcu when OBS starts as described [here](https://github.com/WarmUpTill/SceneSwitcher/wiki/Start-other-programs-when-starting-OBS).
3. For "Run" enter the location of `obs-mcu.exe`
4. Press the "+" under "Arguments" and add the argument `-x`

Now obs-mcu will start whenever OBS starts and automatically connect to both OBS and your MCU. Note that when OBS crashes the app can't detect the shutdown and will stay opened. You will have to quit it using the System Tray in that case.

#### Systray

The app has a systray icon that allows you to quit the app and to open the config file. It also allows you to select the MIDI in and out ports. Restart the app to apply any changes.

### Caveats

Handling of MIDI device disconnects is currently not very graceful, it might take a while until the app detects changes in the MIDI setup.

Theres afaict no way to get the "hidden" state of audio channels, so they will always display on the MCU even if they're hidden in OBS. As a workaround you can simply name these channels so that they appear all the way on the left and then press the "channel right" button until all channels you don't want to see are hidden.

### TODO / Future

Heres a list of planned features I might get around to work on but some of them depend on features that are not yet available in OBS-websockets.

##### Features

- [ ] Allow loading different config files
- [ ] Video fade on master fader

### Development

Building should be straightforward, on linux you need to have libasound2-dev libgtk-3-dev and libayatana-appindicator3-dev available.

- Install golang
- On Linux run `sudo apt-get install clang gcc libasound2-dev libgtk-3-dev libayatana-appindicator3-dev`
- Run `go build` or use the Makefile with `make`

##### Overview

Theres basically two runloops, one for the connection to the MCU and one for the connection to OBS. The latter is handling most of the logic while the MCU loop is basically just translating and trying to keep the amount of data being sent via MIDI low. They communicate through two channels via Message structs and keep draining each others messages while communicating with the MCU respectively OBS.

Having just two runloops instead of a heap of go routines makes it easy to track the control flow logic so theres no need for excessive locking and backchecking.

### Thanks

Thanks to [chabad360](https://github.com/chabad360) for his [gomcu](https://github.com/chabad360/gomcu) code which is integrated in this app.
