package services

import (
	"context"
	pb "paydex/pkg/gen"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) InitStkPush(ctx context.Context, in *pb.StkPushRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
