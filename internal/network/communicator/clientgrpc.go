package communicator

import (
	"context"
	"log"

	"parabellum.kproducer/internal/model"

	pb "github.com/ITA-Dnipro/Dp-230-Result-Collector/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientGRPC struct {
	client     pb.ReportServiceClient
	connection *grpc.ClientConn
}

func NewClientGRPC(serverAddr string) *ClientGRPC {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panicf("cannot connect to grpc server on\t%s, because of\t%v\n", serverAddr, err)
	}

	return &ClientGRPC{
		client:     pb.NewReportServiceClient(conn),
		connection: conn,
	}
}

func (gr *ClientGRPC) CreateNewTask(ctx context.Context, taskData model.TaskFromAPI) (model.TaskProduce, error) {
	result := model.TaskProduce{
		TaskFromAPI: taskData,
		ID:          "",
	}

	request := &pb.CreateReq{URL: taskData.URL, Email: taskData.Email, TotalTestCount: int64(len(taskData.ForwardTo))}
	response, err := gr.client.Create(ctx, request)
	if err != nil {
		return result, err
	}
	result.ID = response.Report.ID

	return result, nil
}

func (gr *ClientGRPC) GetReport(ctx context.Context, id string) (*model.Report, error) {
	req := &pb.GetReportReq{ID: id}
	res, err := gr.client.GetReport(ctx, req)
	if err != nil {
		return nil, err
	}
	return ReportFromProto(res.GetReport()), nil
}

func ReportFromProto(report *pb.Report) *model.Report {
	rep := &model.Report{
		URL: report.URL,
	}
	for _, tr := range report.GetTestResults() {
		rep.TestResults = append(rep.TestResults, TestResultFromProto(tr))
	}
	return rep
}

func TestResultFromProto(tr *pb.TestResult) model.TestResult {
	r := model.TestResult{
		Type: tr.Type,
	}
	for _, res := range tr.GetResults() {
		r.Results = append(r.Results, ResultFromProto(res))
	}
	return r
}

func ResultFromProto(res *pb.Result) model.Result {
	r := model.Result{
		URL:       res.GetURL(),
		StartTime: res.GetStartTime().AsTime(),
		EndTime:   res.GetEndTime().AsTime(),
	}
	for _, poc := range res.GetPoCs() {
		r.PoCs = append(r.PoCs, PoCFromProto(poc))
	}
	return r
}

func PoCFromProto(p *pb.PoC) model.PoC {
	return model.PoC{
		Type:       p.GetType(),
		InjectType: p.GetInjectType(),
		PoCType:    p.GetPoCType(),
		Method:     p.GetMethod(),
		Data:       p.GetData(),
		Param:      p.GetParam(),
		Payload:    p.GetPayload(),
		Evidence:   p.GetEvidence(),
		CWE:        p.GetSWE(),
		Severity:   p.GetSeverity(),
	}
}

func (gr *ClientGRPC) Close() error {
	log.Println("Closing grpc-client")

	return gr.connection.Close()
}
