package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/pb"
	"github.com/the-eduardo/Go-Bank/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	if violations := validateVerifyEmailRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailID:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}
	resp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return resp, nil
}
func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	// Deny if GetId is 0 or has a - value
	if req.GetEmailId() <= 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "id",
			Description: "id is invalid",
		})
	}
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal().Msgf("cannot load config: %v", err)
	}
	if req.GetSecretCode() == "" || len(req.GetSecretCode()) != config.SecretCodeLength {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "secret_code",
			Description: "secret code is invalid",
		})
	}
	return violations
}
