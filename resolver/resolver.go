package resolver

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"kilogram-api/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*model.State
}

func NewRootResolver() *Resolver {
	spamChat := &model.Chat{
		ID: model.SpamChatID,

		Type:  model.ChatTypeGroup,
		Image: &model.SpamChatImage,
		Name:  model.SpamChatName,

		CreatedAt: time.Now(),
		Observers: make(map[string]*model.ChatObserver),

		AllMembersByLogin: make(map[string]*model.User),
		AllMessagesByID:   make(map[string]*model.Message),
	}

	return &Resolver{
		State: &model.State{
			Chats:     []*model.Chat{spamChat},
			ChatsByID: map[string]*model.Chat{spamChat.ID: spamChat},

			UsersByLogin: make(map[string]*model.User),
		},
	}
}

func (r *Resolver) DumpState() {
	file, err := os.OpenFile("state.json", os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		log.Println(err)

		return
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")

	r.ChatsMu.RLock()
	defer r.ChatsMu.RUnlock()

	r.UsersMu.RLock()
	defer r.UsersMu.RUnlock()

	if err := enc.Encode(&r.State); err != nil {
		log.Println(err)

		return
	}

	log.Println("serialize state to json success")
}

func (r *Resolver) LoadState() {
	file, err := os.Open("state.json")

	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		log.Println(err)

		return
	}

	enc := json.NewDecoder(file)

	state := &model.State{}

	if err := enc.Decode(state); err != nil {
		log.Println(err)

		return
	}

	state.UsersByLogin = make(map[string]*model.User, len(state.Users))

	for _, user := range state.Users {
		state.UsersByLogin[user.Login] = user
	}

	state.ChatsByID = make(map[string]*model.Chat, len(state.Chats))

	for _, chat := range state.Chats {
		state.ChatsByID[chat.ID] = chat

		if chat.OwnerLogin != nil {
			chat.Creator = state.UsersByLogin[*chat.OwnerLogin]
		}

		chat.AllMembersByLogin = make(map[string]*model.User, len(chat.AllMembers))

		for _, member := range chat.AllMembers {
			chat.AllMembersByLogin[member.Login] = member
		}

		chat.AllMessagesByID = make(map[string]*model.Message, len(chat.AllMessages))

		for _, message := range chat.AllMessages {
			chat.AllMessagesByID[message.ID] = message

			if message.AuthorLogin != nil {
				message.Author = state.UsersByLogin[*message.AuthorLogin]
			}
		}

		chat.Observers = make(map[string]*model.ChatObserver)
		chat.MessageID = uint64(len(chat.AllMessages))
	}

	state.ChatID = uint64(len(state.Chats))

	r.State = state

	log.Println("deserialize state from json success")
}
