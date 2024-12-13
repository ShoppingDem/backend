// @title Shop API
// @version 1.0
// @description This is a sample server for a shop application.
// @host localhost:8080
// @BasePath /
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ShoppingDem/backend/shop/internal/auth"
	"github.com/ShoppingDem/backend/shop/internal/database"
	"github.com/ShoppingDem/backend/shop/internal/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
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

	// Initialize Okta client
	if err := auth.InitOktaClient(); err != nil {
		log.Fatalf("failed to initialize Okta client: %v", err)
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: db}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
