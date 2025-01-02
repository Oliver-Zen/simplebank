package api

import (
	"fmt"

	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/Oliver-Zen/simplebank/token"
	"github.com/Oliver-Zen/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// `Server` servers HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// `NewServer` creaes a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	// register our custom validator with Gin.
	// `binding.Validator.Engine()` gets the current validator engine the gin is using.
	// `(*validator.Validate)` converts output to a validator.Validate pointer so we can access its methods.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

// SetupRouter adds routes to `router`.
func (server *Server) setupRouter() {
	router := gin.Default()

	// "register new API in the server to route request to handler"
	
	// create user & login user don't need authorization
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	// all other APIs must be protected by authMiddlware
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	
	authRoutes.POST("/accounts", server.createAccount) // last func is the real handlers, others are middleware
	authRoutes.GET("/accounts/:id", server.getAccount) // `:` tells Gin `id` is a URI parameter
	authRoutes.GET("/accounts", server.listAccount)
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

// `Start` runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	// WHY make `Start` public? <- `router`is private
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	// `H` is shortcut for `map[string]any`
	return gin.H{"error": err.Error()}
}
