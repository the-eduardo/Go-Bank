package gapi

import (
	"fmt"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/pb"
	"github.com/the-eduardo/Go-Bank/token"
	"github.com/the-eduardo/Go-Bank/util"
)

// Server serves gRPC requests
type Server struct {
	pb.UnimplementedGoBankServer
	config     util.Config
	tokenMaker token.Maker
	store      db.Store
}

// NewServer creates a new gRPC server
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
	return server, nil

}
