package gapi

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/pb"
	"github.com/the-eduardo/Go-Bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}
	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, status.Errorf(codes.AlreadyExists, "user already exists: %s", err)
			case "23503":
				return nil, status.Errorf(codes.Unavailable, "foreign key violation: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)

	}
	resp := &pb.CreateUserResponse{
		User: convertUser(user),
	}
	return resp, nil
}
