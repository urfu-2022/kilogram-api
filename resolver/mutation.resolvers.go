package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"kilogram-api/model"
	"kilogram-api/server"
	"sync/atomic"
	"time"
)

func (r *mutationResolver) Register(ctx context.Context, login string, password string, name string) (*model.User, error) {
	r.UsersMu.Lock()
	defer r.UsersMu.Unlock()

	if _, ok := r.UsersByLogin[login]; ok {
		return nil, ErrUserAlreadyExists
	}

	user := &model.User{Login: login, Name: name, Password: password}

	r.Users = append(r.Users, user)
	r.UsersByLogin[login] = user

	return user, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, image *string, name *string) (*model.User, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	if image != nil {
		user.M.Lock()
		defer user.M.Unlock()

		if err := validateBase64(*image); err != nil {
			return nil, err
		}

		user.Image = image
	}

	if name != nil {
		user.M.Lock()
		defer user.M.Unlock()

		user.Name = *name
	}

	return user, nil
}

func (r *mutationResolver) UpsertUserMeta(ctx context.Context, key string, val string) (*model.User, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	user.M.Lock()
	defer user.M.Unlock()

	user.Meta = appendMeta(user.Meta, key, val)

	return user, nil
}

func (r *mutationResolver) CreateChat(ctx context.Context, typeArg model.ChatType, name string, members []string) (*model.Chat, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	uniqMembers := map[string]*model.User{user.Login: user}

	r.UsersMu.RLock()
	for _, login := range members {
		if member, ok := r.UsersByLogin[login]; ok {
			uniqMembers[login] = member
		}
	}
	r.UsersMu.RUnlock()

	if (typeArg == model.ChatTypeChannel || typeArg == model.ChatTypeGroup) && len(uniqMembers) < 3 {
		return nil, ErrGroupChatSize
	}

	if typeArg == model.ChatTypePrivate && len(uniqMembers) != 2 {
		return nil, ErrPrivateChatSize
	}

	chatMembers := make([]*model.User, 0, len(uniqMembers))

	for _, member := range uniqMembers {
		chatMembers = append(chatMembers, member)
	}

	chat := &model.Chat{
		ID:   fmt.Sprint(atomic.AddUint64(&r.ChatID, 1)),
		Type: typeArg,

		Name: name,

		AllMembers:        chatMembers,
		AllMembersByLogin: uniqMembers,

		AllMessagesByID: make(map[string]*model.Message),

		Creator:    user,
		OwnerLogin: &user.Login,

		CreatedAt: time.Now(),

		Observers: make(map[string]*model.ChatObserver),
	}

	r.ChatsMu.Lock()
	defer r.ChatsMu.Unlock()

	r.Chats = append(r.Chats, chat)
	r.ChatsByID[chat.ID] = chat

	return chat, nil
}

func (r *mutationResolver) InviteUser(ctx context.Context, chatID string, login string) (bool, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return false, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return false, ErrChatDoesnotExists
	}

	if chat.Type == model.ChatTypePrivate {
		return false, ErrPrivateChatSize
	}

	r.UsersMu.RLock()
	member, ok := r.UsersByLogin[login]
	r.UsersMu.RUnlock()

	if !ok {
		return false, ErrUserDoesnotExists
	}

	if chat.Creator != user {
		return false, ErrNotAuthorized
	}

	if _, ok := chat.AllMembersByLogin[login]; ok {
		return false, ErrAlreadyInvited
	}

	chat.AllMembersByLogin[login] = member
	chat.AllMembers = append(chat.AllMembers, member)

	return true, nil
}

func (r *mutationResolver) KickUser(ctx context.Context, chatID string, login string) (bool, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return false, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return false, ErrChatDoesnotExists
	}

	if chat.Type == model.ChatTypePrivate {
		return false, ErrPrivateChatSize
	}

	if (chat.Type == model.ChatTypeChannel || chat.Type == model.ChatTypeGroup) && len(chat.AllMembers) < 4 {
		return false, ErrGroupChatSize
	}

	if chat.Creator.Login == login {
		return false, ErrKickingYourself
	}

	if chat.Creator != user {
		return false, ErrNotAuthorized
	}

	if _, ok := chat.AllMembersByLogin[login]; !ok {
		return false, ErrNotInvited
	}

	index := -1

	for i, member := range chat.AllMembers {
		if member.Login == login {
			index = i

			break
		}
	}

	if index != -1 {
		delete(chat.AllMembersByLogin, login)
		chat.AllMembers = append(chat.AllMembers[:index], chat.AllMembers[index+1:]...)

		return true, nil
	}

	return true, nil
}

func (r *mutationResolver) UpdateChat(ctx context.Context, id string, image *string, name *string) (*model.Chat, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[id]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	if chat.Creator != user {
		return nil, ErrNotAuthorized
	}

	if image != nil {
		chat.M.Lock()
		defer chat.M.Unlock()

		if err := validateBase64(*image); err != nil {
			return nil, err
		}

		chat.Image = image
	}

	if name != nil {
		chat.M.Lock()
		defer chat.M.Unlock()

		chat.Name = *name
	}

	return chat, nil
}

func (r *mutationResolver) UpsertChatMeta(ctx context.Context, id string, key string, val string) (*model.Chat, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[id]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	if chat.Creator != user {
		return nil, ErrNotAuthorized
	}

	chat.M.Lock()
	defer chat.M.Unlock()

	chat.Meta = appendMeta(chat.Meta, key, val)

	return chat, nil
}

func (r *mutationResolver) DeleteChat(ctx context.Context, id string) (bool, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return false, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[id]
	r.ChatsMu.RUnlock()

	if !ok {
		return false, nil
	}

	if chat.Creator != user {
		return false, ErrNotAuthorized
	}

	r.ChatsMu.Lock()
	defer r.ChatsMu.Unlock()

	index := -1

	for i, chat := range r.Chats {
		if chat.ID == id {
			index = i
		}
	}

	if index != -1 {
		delete(r.ChatsByID, id)
		r.Chats = append(r.Chats[:index], r.Chats[index+1:]...)

		return true, nil
	}

	return false, nil
}

func (r *mutationResolver) SendMessage(ctx context.Context, chatID string, text string) (*model.Message, error) {
	user := GetCurrentUserFrom(ctx)

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	var authorLogin *string

	if user != nil {
		authorLogin = &user.Login
	}

	if user != nil && chat.ID != model.SpamChatID {
		if _, ok := chat.AllMembersByLogin[user.Login]; !ok {
			return nil, ErrMembership
		}
	}

	if chat.Type == model.ChatTypeChannel && user != chat.Creator {
		return nil, model.ErrUnauthorized
	}

	message := &model.Message{
		ID: fmt.Sprint(atomic.AddUint64(&chat.MessageID, 1)),

		Author:      user,
		AuthorLogin: authorLogin,

		CreatedAt: time.Now(),

		Text: text,
	}

	chat.M.Lock()
	chat.AllMessages = append(chat.AllMessages, message)
	chat.AllMessagesByID[message.ID] = message
	chat.M.Unlock()

	for _, observer := range chat.Observers {
		if user == nil || message.Author == nil || observer.Login != message.Author.Login {
			observer.Message <- message
		}
	}

	return message, nil
}

func (r *mutationResolver) EditMessage(ctx context.Context, chatID string, messageID string, text string) (*model.Message, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	message, ok := chat.AllMessagesByID[messageID]

	if !ok {
		return nil, ErrMessageDoesnotExists
	}

	if user != message.Author {
		return nil, ErrNotAuthorized
	}

	message.Text = text

	return message, nil
}

func (r *mutationResolver) UpsertMessageMeta(ctx context.Context, chatID string, messageID string, key string, val string) (*model.Message, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return nil, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return nil, ErrChatDoesnotExists
	}

	message, ok := chat.AllMessagesByID[messageID]

	if !ok {
		return nil, ErrMessageDoesnotExists
	}

	message.Meta = appendMeta(message.Meta, key, val)

	return message, nil
}

func (r *mutationResolver) DeleteMessage(ctx context.Context, chatID string, messageID string) (bool, error) {
	user := GetCurrentUserFrom(ctx)

	if user == nil {
		return false, ErrNotAuthorized
	}

	r.ChatsMu.RLock()
	chat, ok := r.ChatsByID[chatID]
	r.ChatsMu.RUnlock()

	if !ok {
		return false, ErrChatDoesnotExists
	}

	index := -1

	for i, message := range chat.AllMessages {
		if message.ID == messageID {
			index = i

			if message.Author != user && chat.Creator != user {
				return false, ErrNotAuthorized
			}

			break
		}
	}

	if index != -1 {
		delete(chat.AllMessagesByID, messageID)
		chat.AllMessages = append(chat.AllMessages[:index], chat.AllMessages[index+1:]...)

		return true, nil
	}

	return false, nil
}

// Mutation returns server.MutationResolver implementation.
func (r *Resolver) Mutation() server.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
