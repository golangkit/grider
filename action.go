package grider

import (
	"encoding/json"
	"errors"
)

// Action describes an operation on UI user can invocate.
// An action could relate to the specific row or not.
type Action struct {
	Code       ActionCode  `json:"code,omitempty"`
	Perm       string      `json:"perm,omitempty"`
	Title      string      `json:"title,omitempty"`
	Icon       *Icon       `json:"icon,omitempty"`
	DirectCall *DirectCall `json:"directCall,omitempty"`
}

// DirectCall is action's extended attributes describing a simple action
// what does not have user interaction. As instance an action what directly calls
// server REST handler.
type DirectCall struct {
	IsConfirmationRequired bool            `json:"isConfirmationRequired,omitempty"`
	ConfirmationMessage    string          `json:"confirmationMessage,omitempty"`
	Method                 string          `json:"method,omitempty"`
	Path                   string          `json:"path,omitempty"`
	Body                   json.RawMessage `json:"body,omitempty"`
}

type ActionCode string

// ActionSet holds Actions identified by code.
// The type is useful for describe all supported application actions.
type ActionSet map[ActionCode]Action

// NewActionSet builds instance of ActionSet.
func NewActionSet() map[ActionCode]Action {
	return map[ActionCode]Action{}
}

func (as ActionSet) Add(arr []ActionCode) {
	if len(arr) == 0 {
		return
	}

	for i := range arr {
		as[arr[i]] = Action{}
	}
}

func (dst ActionSet) AssignActionValues(src ActionSet) error {
	for k := range dst {
		v, ok := src[k]
		if !ok {
			return errors.New("unknown action code:" + string(k))
		}
		dst[k] = v
	}
	return nil
}
