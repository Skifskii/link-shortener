package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Skifskii/link-shortener/internal/repository"
	"github.com/Skifskii/link-shortener/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	proto.UnimplementedShortenerServiceServer
	shortener Shortener
}

// Shortener интерфейс для сервиса сокращения ссылок.
type Shortener interface {
	Shorten(userID int, longURL string) (shortURL string, err error)
}

func New(shortener Shortener) *GRPCServer {
	return &GRPCServer{
		shortener: shortener,
	}
}

func (g *GRPCServer) Run(port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	// Запускаем сервер
	fmt.Printf("Starting gRPC server at %s\n", listen.Addr().String())

	s := grpc.NewServer()
	proto.RegisterShortenerServiceServer(s, g)

	return s.Serve(listen)
}

func (g *GRPCServer) ShortenURL(ctx context.Context, in *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	var resp proto.URLShortenResponse

	userID := 0 // TODO: научиться получать userID

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
