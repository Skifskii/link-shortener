package interceptors

import (
	"context"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auther - интерфейс для работы с пользователями и JWT-токенами.
type Auther interface {
	// CreateUser создает нового пользователя и возвращает его JWT-токен.
	CreateUser(username string) (jwt string, err error)
	// GetUserID извлекает идентификатор пользователя из JWT-токена.
	GetUserID(tokenString string) (int, error)
}

func NewAuthUnaryInterceptor(a Auther) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var headerAuth string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				headerAuth = values[0]
			}
		}

		var token string
		if len(headerAuth) == 0 {
			token, err = a.CreateUser("")
			if err != nil {
				return nil, status.Error(codes.Internal, "ошибка при создании токена")
			}
		} else {
			token = headerAuth
		}

		userID, err := a.GetUserID(token)
		if err != nil {
			return nil, status.Error(codes.Internal, "ошибка при определении userID")

		}

		ctx = context.WithValue(ctx, authmw.UserIDKey, userID)

		return handler(ctx, req)
	}
}
