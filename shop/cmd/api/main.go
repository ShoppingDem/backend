package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ShoppingDem/backend/shop/internal/database"
	"github.com/ShoppingDem/backend/shop/internal/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Remove the Okta client initialization
	// if err := auth.InitOktaClient(); err != nil {
	// 	log.Fatalf("failed to initialize Okta client: %v", err)
	// }

	// Create the base server.
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))

	// 1. Configure transports (order matters here):
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin (adjust as needed for security).
				return true
			},
			HandshakeTimeout: 5 * time.Second, // Customize timeout
		},
		KeepAlivePingInterval: 10 * time.Second,    // Keep-alive ping
		InitFunc:              graph.WebsocketInit, // Add an init function for Websockets
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// 2. Add extensions (e.g., for complexity limit, tracing, etc.):
	srv.Use(extension.Introspection{}) // Enable introspection queries (useful for development)
	srv.Use(extension.AutomaticPersistedQueries{})
	// srv.Use(extension.FixedComplexityLimit(100)) // Set a complexity limit (adjust as needed)

	// 3. Error handling
	// You can add a custom error presenter or formatter here if you need to customize error responses.
	// srv.SetErrorPresenter(...)
	// srv.SetRecoverFunc(...)

	// 4. Query Complexity
	// You can implement more advanced query complexity calculation if necessary.

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
