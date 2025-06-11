package main

// inspiration: https://github.com/matt-slater/csi-driver/tree/main/cmd

/* TODO:
 * cli flags instead of env vars
 */

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"

	"github.com/gulducat/csi-madrid/pkg/server"
	"github.com/gulducat/csi-madrid/pkg/sink"
)

func main() {
	csiSock := os.Getenv("CSI_ENDPOINT")
	if csiSock == "" {
		log.Fatal("require CSI_ENDPOINT path to unix socket")
	}
	nodeID := os.Getenv("NODE_ID")

	logger := hclog.New(&hclog.LoggerOptions{
		Name: "csi-madrid",
		// Level: hclog.Debug,
		Level: hclog.Info,
		// IncludeLocation: true,
		DisableTime: true, // csi logging already has a stamp
	})

	volSink, err := newSink(logger)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.Remove(csiSock); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	l, err := net.Listen("unix", csiSock)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", csiSock, err)
	}

	logger.Info("starting server", "sock", csiSock)
	serv := server.New(logger, nodeID, volSink)
	errCh := make(chan error)
	go func() {
		errCh <- serv.Serve(l)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Kill, os.Interrupt)

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("error from grpc server", "error", err)
		} else {
			logger.Info("done with server")
		}
	case s := <-sig:
		logger.Info("done", "signal", s)
		serv.GracefulStop()
	}
}

func newSink(log hclog.Logger) (sink.Sink, error) {
	// sink, err := sink.NewFileSink(logger, "/tmp/csi-madrid") // TODO: parameterize
	// if err != nil {
	// 	log.Fatalf("error setting up sink: %s", err)
	// }
	nomadPath := os.Getenv("NOMAD_SINK_PATH")
	if nomadPath != "" {
		log.Info("using nomad sink")
		conf := api.DefaultConfig()
		client, err := api.NewClient(conf)
		if err != nil {
			return nil, fmt.Errorf("error getting nomad client: %w", err)
		}
		return sink.NewNomadSink(
			log.Named("nomad-sink"),
			client.Variables(), nomadPath,
			conf.Region, conf.Namespace), nil
	}
	log.Info("using memory sink")
	return sink.NewMemSink(), nil
}
