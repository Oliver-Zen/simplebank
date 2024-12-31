package api

import (
	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// `Server` servers HTTP requests for our banking service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// `NewServer` creaes a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// register our custom validator with Gin.
	// `binding.Validator.Engine()` gets the current validator engine the gin is using.
	// `(*validator.Validate)` converts output to a validator.Validate pointer so we can access its methods.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	// "register new API in the server to route request to handler"
	// add routes to router
	router.POST("/users", server.createUser)
	router.POST("/accounts", server.createAccount) // last func is the real handlers, others are middleware
	router.GET("/accounts/:id", server.getAccount) // `:` tells Gin `id` is a URI parameter
	router.GET("/accounts", server.listAccount)
	router.POST("/transfers", server.createTransfer) 

	server.router = router
	return server
}

// `Start` runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	// WHY make `Start` public? <- `router`is private
	return server.router.Run(address)
}

// `H` is shortcut for `map[string]any`
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
