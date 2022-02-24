package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"kilogram-api/model"
	"kilogram-api/server"
	"math/rand"
)

func (r *subscriptionResolver) NewEvent(ctx context.Context) (<-chan model.Event, error) {
	if GetCurrentUserFrom(ctx) == nil {
		return nil, ErrNotAuthorized
	}

	events := make(chan model.Event, 1)

	r.ChatsMu.RLock()
	for _, chat := range r.Chats {
		chat := chat

		go func() {
			newMessageChan, err := r.NewMessage(ctx, chat.ID)

			if err != nil {
				return
			}

			for message := range newMessageChan {
				events <- model.MessageEvent{Chat: chat, Message: message}
			}
		}()
	}
	r.ChatsMu.RUnlock()

	return events, nil
}

func (r *subscriptionResolver) NewMessage(ctx context.Context, chatID string) (<-chan *model.Message, error) {
	var login string

	user := GetCurrentUserFrom(ctx)

	if user != nil {
		login = user.Login
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	if _, ok := chat.AllMembersByLogin[login]; !ok {
		return nil, ErrNotAuthorized
	}

	observerID := string(rand.Int31()) // nolint: gosec
	events := make(chan *model.Message, 1)

	go func() {
		<-ctx.Done()

		chat.M.Lock()
		defer chat.M.Unlock()

		delete(chat.Observers, observerID)
	}()

	chat.M.Lock()
	chat.Observers[observerID] = &model.ChatObserver{Login: login, Message: events}
	chat.M.Unlock()

	return events, nil
}

// Subscription returns server.SubscriptionResolver implementation.
func (r *Resolver) Subscription() server.SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }
