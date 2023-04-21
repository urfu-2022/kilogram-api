package server

import (
	"context"
	"errors"
	"kilogram-api/model"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

type ctxKeySignature int

const signatureKey ctxKeySignature = 0

func WithSignature(ctx context.Context, signature string) context.Context {
	return context.WithValue(ctx, signatureKey, signature)
}

func GetSignatureFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if result, ok := ctx.Value(signatureKey).(string); ok {
		return result
	}

	return ""
}

func New(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.Use(extension.Introspection{})

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			HandshakeTimeout: time.Minute,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			EnableCompression: true,
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			signature := initPayload.Authorization()
			_, err := model.ValidateUser(signature)

			if err != nil {
				return ctx, errors.New("AUTHORIZATION_REQUIRED")
			}

			return WithSignature(ctx, signature), nil
		},
	})

	return srv
}
