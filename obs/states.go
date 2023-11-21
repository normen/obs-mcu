package obs

import (
	"reflect"
	"strings"

	"github.com/normen/obs-mcu/config"
	"github.com/normen/obs-mcu/msg"
)

type ObsState struct {
	StateName string
	LedName   string
	State     bool
}

type ObsStates struct {
	states map[string]*ObsState
}

func NewObsStates() *ObsStates {
	ret := &ObsStates{
		states: make(map[string]*ObsState),
	}
	ret.getConfig()
	return ret
}

func (s *ObsStates) SetState(name string, state bool) {
	if st, ok := s.states[name]; ok {
		if st.State != state {
			st.State = state
			fromObs <- msg.LedMessage{
				LedName:  st.LedName,
				LedState: st.State,
			}
		}
	}
}

func (s *ObsStates) SendAll() {
	for _, st := range s.states {
		fromObs <- msg.LedMessage{
			LedName:  st.LedName,
			LedState: st.State,
		}
	}
}

func (s *ObsStates) GetState(name string) *ObsState {
	if st, ok := s.states[name]; ok {
		return st
	}
	return nil
}

func (s *ObsStates) DeleteState(name string, state bool) {
	delete(s.states, name)
}

func (t *ObsStates) getConfig() {
	s := reflect.ValueOf(config.Config.McuLeds).Elem()
	num := s.NumField()
	for i := 0; i < num; i++ {
		fieldVal := s.FieldByIndex([]int{i})
		if fieldVal.IsValid() && fieldVal.Kind() == reflect.String {
			ledName := reflect.TypeOf(*config.Config.McuLeds).Field(i).Name
			configVal := fieldVal.String()
			if configVal != "" {
				ledType, stateName, found := strings.Cut(configVal, ":")
				if found {
					switch ledType {
					case "STATE":
						st := ObsState{
							StateName: stateName,
							LedName:   ledName,
							State:     false,
						}
						t.states[stateName] = &st
					}
				}
			}
		}
	}
}
