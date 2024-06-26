package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"kilogram-api/resolver"
	"kilogram-api/server"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
)

var (
	endpoint     = ":8080"
	dumpInterval = 5 * time.Minute
)

func main() {
	if port := os.Getenv("PORT"); port != "" {
		endpoint = fmt.Sprintf(":%s", port)
	}

	resolver := resolver.NewRootResolver()
	config := server.Config{Resolvers: resolver}
	schema := server.NewExecutableSchema(config)
	srv := server.New(schema)

	resolver.LoadState()

	ticker := time.NewTicker(dumpInterval)

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				resolver.DumpState()

			case <-quit:
				ticker.Stop()

				return
			}
		}
	}()

	router := chi.NewRouter()
	router.Get("/", playground.Handler("GraphQL playground", "/query"))
	router.Get("/static/{file}", server.GetFile)
	router.Post("/upload", server.UploadFile)
	router.With(server.CORS, resolver.CurrentUserMiddleware).Handle("/query", srv)

	log.Printf("running on %s", endpoint)
	log.Println(http.ListenAndServe(endpoint, router))

	close(quit)
}
