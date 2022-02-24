package resolver

import (
	"context"
	"crypto/subtle"
	"net/http"

	"kilogram-api/model"
)

type ctxKeyUser int

const userKey ctxKeyUser = 0

func (r *Resolver) CurrentUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		signature := request.Header.Get("authorization")
		claims, err := model.ValidateUser(signature)
		runNext := func() { next.ServeHTTP(response, request.WithContext(ctx)) }

		if err != nil {
			runNext()

			return
		}

		r.UsersMu.RLock()
		user, ok := r.UsersByLogin[claims.Login]
		r.UsersMu.RUnlock()

		if !ok {
			runNext()

			return
		}

		if subtle.ConstantTimeCompare([]byte(claims.Password), []byte(user.Password)) != 0 {
			ctx = context.WithValue(ctx, userKey, user)
		}

		runNext()
	})
}

func GetCurrentUserFrom(ctx context.Context) *model.User {
	if ctx == nil {
		return nil
	}

	if result, ok := ctx.Value(userKey).(*model.User); ok {
		return result
	}

	return nil
}
