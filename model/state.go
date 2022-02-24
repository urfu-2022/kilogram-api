package model

import (
	"sync"
)

type State struct {
	ChatID uint64 `json:"-"`

	Chats []*Chat `json:"chats"`
	Users []*User `json:"users"`

	ChatsByID    map[string]*Chat `json:"-"`
	UsersByLogin map[string]*User `json:"-"`

	ChatsMu sync.RWMutex `json:"-"`
	UsersMu sync.RWMutex `json:"-"`
}
