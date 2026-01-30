package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Skifskii/link-shortener/internal/middleware/authmw"
	"github.com/Skifskii/link-shortener/internal/model"
	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/internal/transport/grpc/interceptors"
	"github.com/Skifskii/link-shortener/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	proto.UnimplementedShortenerServiceServer
	shortener Shortener
	auth      Auther
	sr        ShortRedirecter
	u         UserPairsGetter
}

// Shortener интерфейс для сервиса сокращения ссылок.
type Shortener interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
}

// Auther - интерфейс для работы с пользователями и JWT-токенами.
type Auther interface {
	// CreateUser создает нового пользователя и возвращает его JWT-токен.
	CreateUser(username string) (jwt string, err error)
	// GetUserID извлекает идентификатор пользователя из JWT-токена.
	GetUserID(tokenString string) (int, error)
}

// ShortRedirecter интерфейс предоставляет метод для получения оригинального URL
// по короткой части ссылки (без baseURL).
type ShortRedirecter interface {
	Redirect(shortURL string) (longURL string, err error)
}

// UserPairsGetter интерфейс для получения списка пар short->original пользователя.
type UserPairsGetter interface {
	GetUserPairs(userID int) ([]model.ResponsePairElement, error)
}

func New(shortener Shortener, auth Auther, sr ShortRedirecter, u UserPairsGetter) *GRPCServer {
	return &GRPCServer{
		shortener: shortener,
		auth:      auth,
		sr:        sr,
		u:         u,
	}
}

func (g *GRPCServer) Run(port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	// Запускаем сервер
	fmt.Printf("Starting gRPC server at %s\n", listen.Addr().String())

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptors.NewAuthUnaryInterceptor(g.auth)))

	proto.RegisterShortenerServiceServer(s, g)

	return s.Serve(listen)
}

func (g *GRPCServer) ShortenURL(ctx context.Context, in *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	var resp proto.URLShortenResponse

	userID, ok := ctx.Value(authmw.UserIDKey).(int)
	if !ok {
		return nil, status.Error(codes.Internal, "ошибка при определении user_id")
	}

	// Сокращаем ссылку
	shortURL, err := g.shortener.Shorten(userID, in.GetUrl())
	if err != nil {
		if errors.Is(err, repository.ErrOriginalURLAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp.SetResult(shortURL)

	return &resp, nil
}

func (g *GRPCServer) ExpandURL(ctx context.Context, in *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	var resp proto.URLExpandResponse

	shortURL := in.GetId()
	if shortURL == "" {
		return nil, status.Error(codes.Internal, "id не указан")
	}

	longURL, err := g.sr.Redirect(shortURL)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	resp.SetResult(longURL)

	return &resp, nil
}

func (g *GRPCServer) ListUserURLs(ctx context.Context, in *emptypb.Empty) (*proto.UserURLsResponse, error) {
	var resp proto.UserURLsResponse

	userID, ok := ctx.Value(authmw.UserIDKey).(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "ошибка аутентификации")
	}

	pairs, err := g.u.GetUserPairs(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(pairs) == 0 {
		return nil, status.Error(codes.NotFound, "ссылки не найдены")
	}

	urls := make([]*proto.URLData, len(pairs))
	for i, pair := range pairs {
		urls[i] = &proto.URLData{}
		urls[i].SetOriginalUrl(pair.OriginalURL)
		urls[i].SetShortUrl(pair.ShortURL)
	}

	resp.SetUrl(urls)

	return &resp, nil
}
