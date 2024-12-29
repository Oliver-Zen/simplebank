package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
)

// `Server` servers HTTP requests for our banking service.
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// `NewServer` creaes a new HTTP server and setup routing.
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// add routes to router
	router.POST("/accounts", server.createAccount) // last func is the real handlers, others are middleware
	router.GET("/accounts/:id", server.getAccount) // `:` tells Gin `id` is a URI parameter
	router.GET("/accounts", server.listAccount)

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
