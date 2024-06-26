package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"kilogram-api/model"
	"kilogram-api/server"

	"github.com/google/uuid"
)

// NewEvent is the resolver for the newEvent field.
func (r *subscriptionResolver) NewEvent(ctx context.Context) (<-chan model.Event, error) {
	events := make(chan model.Event)

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

// NewMessage is the resolver for the newMessage field.
func (r *subscriptionResolver) NewMessage(ctx context.Context, chatID string) (<-chan *model.Message, error) {
	login := server.GetSignatureFrom(ctx)

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	if _, ok := chat.AllMembersByLogin[login]; !ok && chatID != model.SpamChatID {
		return nil, ErrNotAuthorized
	}

	observerID := uuid.New().String()

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
