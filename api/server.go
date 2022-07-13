package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
)

// Server serves HTTP requests
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and sets up the routes
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)

	router.POST("/transfer", server.createTransfer)
	router.GET("/transfer/:id", server.getTransfer)
	router.GET("/transfer", server.listTransfer)

	router.POST("/entries", server.createEntry)
	router.GET("/entries/:id", server.getEntry)
	router.GET("/entries", server.listEntries)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
