package api

import (
	db "GoBank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server provides the HTTP rest API
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and set up routing
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Add routes for accounts
	router.POST("/accounts", server.createAccount)
	router.DELETE("/accounts/:id", server.deleteAccountRequest)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts/", server.listAccount)

	// Add routes for transfers
	router.POST("/transfers", server.createTransfer)
	router.GET("/transfers/:id", server.getTransfer)
	router.GET("/transfers/", server.listTransfers)

	// Add routes for entries
	router.POST("/entries", server.newEntry)
	router.GET("/entries/:id", server.getEntry)
	router.GET("/entries/", server.listEntries)

	server.router = router
	return server

}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
