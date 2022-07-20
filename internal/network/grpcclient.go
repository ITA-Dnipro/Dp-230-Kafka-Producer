package network

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"parabellum.kproducer/internal/model"
)

//import pb "github.com/ITA-Dnipro/Dp-230-Result-Collector/proto"

type ClientGRPC struct {
	client *grpc.ClientConn
	ctx    context.Context
}

func NewClientGRPC(ctx context.Context, serverAddr string) *ClientGRPC {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("cannot connect to grpc server on\t%s, because of\t%v\n", serverAddr, err)
	}

	return &ClientGRPC{
		client: conn,
		ctx:    ctx,
	}
}

func (gr *ClientGRPC) CreateNewTask(taskData model.TaskFromAPI) (model.TaskProduce, error) {
	result := model.TaskProduce{
		TaskFromAPI: taskData,
		ID:          "",
	}

	//TODO: will complete, when have correct .proto published
	//request := &pb.CreateReq{URL: taskData.URL, Email: taskData.Email, TotalTestCount: len(taskData.ForwardTo)}
	//response, err := gr.client.Create(gr.ctx, request)
	//if err!=nil {
	//	return result, err
	//}
	//result.ID = response.ID

	result.ID = "stub-until-i-have-proto"

	return result, nil
}

func (gr *ClientGRPC) Close() error {
	return gr.client.Close()
}
