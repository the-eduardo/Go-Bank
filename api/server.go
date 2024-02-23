package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/token"
	"github.com/the-eduardo/Go-Bank/util"
)

// Server provides the HTTP rest API
type Server struct {
	config     util.Config
	tokenMaker token.Maker
	store      db.Store
	router     *gin.Engine
}

// NewServer creates a new HTTP server and set up routing
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
	server.setupRouter()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)

	}

	return server, nil

}
func (server *Server) setupRouter() {
	router := gin.Default()

	// Add routes for users
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	// Add routes for accounts
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.DELETE("/accounts/:id", server.deleteAccountRequest)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts/", server.listAccount)

	// Add routes for transfers
	authRoutes.POST("/transfers", server.createTransfer)
	authRoutes.GET("/transfers/:id", server.getTransfer)
	authRoutes.GET("/transfers/", server.listTransfers)

	// Add routes for entries
	authRoutes.POST("/entries", server.newEntry)
	authRoutes.GET("/entries/:id", server.getEntry)
	authRoutes.GET("/entries/", server.listEntries)

	server.router = router
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
