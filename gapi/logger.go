package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

func GrpcLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}
	loggerLevel := log.Info()
	if err != nil {
		loggerLevel = log.Error()
	}
	loggerLevel.Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Msg("received a gRPC request")

	return result, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	Body       []byte
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}
func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{ResponseWriter: res, statusCode: http.StatusOK}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		loggerLevel := log.Info()
		if rec.statusCode != http.StatusOK {
			loggerLevel = log.Error().Bytes("response_body", rec.Body)
		}

		loggerLevel.Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Int("status_code", rec.statusCode).
			Str("status_text", http.StatusText(rec.statusCode)).
			Dur("duration", duration).
			Msg("received a HTTP request")
	})
}
