package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/subtle"
	"kilogram-api/model"
	"kilogram-api/server"
)

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	return GetCurrentUserFrom(ctx), nil
}

func (r *queryResolver) SignIn(ctx context.Context, login string, password string) (*string, error) {
	r.UsersMu.RLock()
	user, ok := r.UsersByLogin[login]
	r.UsersMu.RUnlock()

	if !ok {
		return nil, ErrNotAuthorized
	}

	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) == 0 {
		return nil, ErrNotAuthorized
	}

	sign, err := model.SignUser(login, password)

	if err != nil {
		return nil, ErrNotAuthorized
	}

	return &sign, nil
}

func (r *queryResolver) Chats(ctx context.Context, offset *int, first *int) ([]*model.Chat, error) {
	begin := *offset
	count := *first

	user := GetCurrentUserFrom(ctx)

	r.ChatsMu.RLock()
	defer r.ChatsMu.RUnlock()

	chats := r.Resolver.Chats

	if begin >= len(chats) {
		return nil, nil
	}

	if begin < 0 {
		count += begin
		begin = 0
	}

	if count <= 0 {
		return nil, nil
	}

	if len(chats) < begin+count {
		count = len(chats) - begin
	}

	chats = chats[begin:]

	result := make([]*model.Chat, 0, count)

	for i := 0; len(result) < count && i < len(chats); i++ {
		if chats[i].ID == model.SpamChatID {
			result = append(result, chats[i])

			continue
		}

		if user == nil {
			continue
		}

		if _, ok := chats[i].AllMembersByLogin[user.Login]; ok {
			result = append(result, chats[i])

			continue
		}
	}

	return result, nil
}

func (r *queryResolver) Users(ctx context.Context, offset *int, first *int) ([]*model.User, error) {
	begin := *offset
	count := *first

	r.UsersMu.RLock()
	defer r.UsersMu.RUnlock()

	users := r.Resolver.Users

	if begin >= len(users) {
		return nil, nil
	}

	if begin < 0 {
		count += begin
		begin = 0
	}

	if count < 0 {
		return nil, nil
	}

	if len(users) < begin+count {
		count = len(users) - begin
	}

	return users[begin : begin+count], nil
}

// Query returns server.QueryResolver implementation.
func (r *Resolver) Query() server.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
