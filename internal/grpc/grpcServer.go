package grpcserver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "pvzService/internal/proto"
)

type PVZServer struct {
	pb.UnimplementedPVZServiceServer
	db *sql.DB
}

func NewPVZServer(db *sql.DB) *PVZServer {
	return &PVZServer{db: db}
}

func (s *PVZServer) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	rows, err := s.db.Query(`
        SELECT 
          p.id, p.registration_date, p.city
        FROM pvz p
        ORDER BY p.registration_date
      `)
	if err != nil {
		return nil, fmt.Errorf("failed to query PVZ list: %w", err)
	}
	defer rows.Close()

	var pvzList []*pb.PVZ

	for rows.Next() {
		var pvzID sql.NullString
		var pvzRegDate sql.NullTime
		var pvzCity sql.NullString

		err = rows.Scan(
			&pvzID, &pvzRegDate, &pvzCity,
		)

		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		if pvzID.Valid {
			pvz := &pb.PVZ{
				Id:               pvzID.String,
				RegistrationDate: timestamppb.New(pvzRegDate.Time),
				City:             pvzCity.String,
			}

			pvzList = append(pvzList, pvz)
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
	}

	return &pb.GetPVZListResponse{Pvzs: pvzList}, nil
}

func StartGRPCServer(db *sql.DB, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterPVZServiceServer(s, NewPVZServer(db))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
