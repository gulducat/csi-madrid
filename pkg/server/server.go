package server

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/gulducat/csi-madrid/pkg/sink"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

func New(logger hclog.Logger, nodeID string, sink sink.Sink) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	}
	serv := grpc.NewServer(opts...)
	csi.RegisterIdentityServer(serv, &IdentityServer{
		Name:    "csi-madrid",
		Version: "0.0.1",
	})
	csi.RegisterControllerServer(serv, &ControllerServer{
		sink: sink,
		log:  logger.Named("controller"),
	})
	csi.RegisterNodeServer(serv, &NodeServer{
		nodeID: nodeID,
		log:    logger.Named("node"),
	})
	return serv
}

func loggingInterceptor(logger hclog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		logger.Info("request received",
			"rpc_method", info.FullMethod,
			"request", req,
		)

		resp, err := handler(ctx, req)
		if err != nil {
			logger.Error("failed processing request", "error", err)
		} else {
			if s, ok := resp.(string); ok && s == "" {
				logger.Debug("request completed with empty string response")
				return resp, err
			}
			logger.Info("request completed", "response", resp)
		}

		return resp, err
	}
}
