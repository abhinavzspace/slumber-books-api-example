package main

import (
	"errors"
	"fmt"
	"github.com/abhinavzspace/slumber-books-api-example/books"
	"github.com/abhinavzspace/slumber-books-api-example/hooks"
	"github.com/abhinavzspace/slumber-sessions"
	"github.com/abhinavzspace/slumber-users"
	"github.com/abhinavzspace/slumber/middlewares/context"
	"github.com/abhinavzspace/slumber/middlewares/mongodb"
	"github.com/abhinavzspace/slumber/middlewares/renderer"
	"github.com/abhinavzspace/slumber/server"
	"io/ioutil"
	"time"
)

func main() {

	// try to load signing keys for token authority
	// NOTE: DO NOT USE THESE KEYS FOR PRODUCTION! FOR DEMO ONLY
	privateSigningKey, err := ioutil.ReadFile("keys/demo.rsa")
	if err != nil {
		panic(errors.New(fmt.Sprintf("Error loading private signing key: %v", err.Error())))
	}
	publicSigningKey, err := ioutil.ReadFile("keys/demo.rsa.pub")
	if err != nil {
		panic(errors.New(fmt.Sprintf("Error loading public signing key: %v", err.Error())))
	}

	// create current project context
	ctx := context.New()

	// set up DB session
	db := mongodb.New(&mongodb.Options{
		ServerName:   "localhost",
		DatabaseName: "slumber-books",
	})
	dbSession := db.NewSession()

	// set up Renderer (unrolled_render)
	renderer := renderer.New(&renderer.Options{
		IndentJSON: true,
	}, renderer.JSON)

	// set up users resource
	usersResource := users.NewResource(ctx, &users.Options{
		Renderer: renderer,
		Database: db,
		ControllerHooks: &users.ControllerHooks{
			PostCreateUserHook: hooks.HandlerPostCreateUserHook,
		},
	})

	// set up sessions resource
	sessionsResource := sessions.NewResource(ctx, &sessions.Options{
		PrivateSigningKey:     privateSigningKey,
		PublicSigningKey:      publicSigningKey,
		Renderer:              renderer,
		Database: db,
		UserRepositoryFactory: usersResource.UserRepositoryFactory,
	})

	// set up books resource
	booksResource := books.NewResource(ctx, &books.Options{})

	// init server
	s := server.NewServer(&server.Config{
		Context: ctx,
	})

	// set up router
	ac := server.NewAccessController(ctx, renderer)
	router := server.NewRouter(s.Context, ac)

	// add REST resources to router
	router.AddResources(
		usersResource,
		sessionsResource,
		booksResource,
	)

	// add middlewares
	s.UseContextMiddleware(dbSession)
	s.UseContextMiddleware(renderer)
	s.UseMiddleware(sessionsResource.NewAuthenticator())
	s.UseMiddleware(booksResource)

	// setup router
	s.UseRouter(router)

	// bam!
	s.Run(":3001", server.Options{
		Timeout: 10*time.Second,
	})
}
