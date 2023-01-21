package services

import (
	"context"
	"log"
	pb "paydex/pkg/gen"
	"paydex/worker"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) InitStkPush(ctx context.Context, in *pb.StkPushRequest) (*emptypb.Empty, error) {
	s.l.Info("InitSktPush", in)
	if err := s.worker.DistributeTaskSendSTKPush(ctx, &worker.STKRequest{
		Amount:      in.Amount,
		Description: in.TransactionDesc,
		PhoneNumber: in.PhoneNumber,
	}); err != nil {
		log.Print(err)
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
