package gapi

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/the-eduardo/Go-Bank/api"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/pb"
	"github.com/the-eduardo/Go-Bank/util"
	"github.com/the-eduardo/Go-Bank/val"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	if violations := validateLoginUserRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user")
	}
	err = util.CheckPassword(user.HashedPassword, req.Password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid password")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify password")
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token")
	}
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token")
	}
	fixedRefreshPayload, err := api.ConvertGoogleUUIDToPGTypeUUID(refreshPayload.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot convert refresh payload")
	}
	fixedRefreshPayloadExpiresAt := api.TimeToPGTimestamptz(refreshPayload.ExpiredAt)
	mtdt := server.extractMetadata(ctx)
	_, err = server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           fixedRefreshPayload,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    fixedRefreshPayloadExpiresAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session")
	}

	resp := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             refreshPayload.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}
	return resp, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return violations
}
