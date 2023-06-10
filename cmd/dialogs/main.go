package main

import (
	grpc2 "github.com/antonpriyma/otus-highload/internal/app/dialog/delivery/grpc"
	dialog_repo "github.com/antonpriyma/otus-highload/internal/app/dialog/repository/mysql"
	dialog_usecase "github.com/antonpriyma/otus-highload/internal/app/dialog/usecase"
	"github.com/antonpriyma/otus-highload/pkg/dialogs/github.com/antonpriyma/otus-highload/pkg/dialogs"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoapi"
	"github.com/antonpriyma/otus-highload/pkg/framework/grpc/interceptors/server"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"google.golang.org/grpc"
	"net"
	"time"
)

type AppConfig struct {
	service.Config `mapstructure:",squash"`
	DialogsConfig  DialogsConfig `mapstructure:"dialogs"`
}

type DialogsConfig struct {
	Repo dialog_repo.Config `mapstructure:"repository"`
}

func (a AppConfig) APIConfig() echoapi.Config {
	return echoapi.Config{
		ServeConfig: service.ServeConfig{
			GracefulWait: time.Second,
			StopWait:     time.Second,
		},
		Listen: ":8082",
	}
}

func main() {
	var cfg AppConfig
	cfg.Version = service.Version{
		Dist:    "local",
		Release: "local",
	}

	svc := echoapi.New(&cfg)

	dialogRepo, err := dialog_repo.NewRepository(cfg.DialogsConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create dialogs repository")

	dialogsUsecase := dialog_usecase.NewUsecase(dialogRepo, svc.Logger)
	dialogsGRPCDelivery := grpc2.NewDelivery(dialogsUsecase, svc.Logger)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.NewRequestIDInterceptor(svc.Logger),
			server.NewLoggerStatInterceptor(svc.Logger),
			server.NewAccessLogInterceptor(svc.Logger),
		),
	)
	dialogs.RegisterDialogsServer(grpcServer, dialogsGRPCDelivery)

	lis, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 50051,
	})
	utils.Must(svc.Logger, err, "failed to listen tcp")

	if err := grpcServer.Serve(lis); err != nil {
		svc.Logger.Fatal(err)
	}
}
