package main

// inspiration: https://github.com/matt-slater/csi-driver/tree/main/cmd

/* TODO:
 * optional env vars?
 * not log.Fatal
 */

import (
	"errors"
	"flag"
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
	o, err := getOpts(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:        "csi-madrid",
		Level:       hclog.LevelFromString(o.logLevel),
		DisableTime: true, // csi logging already has a stamp
	})

	logger.Info("setting up data store for volume and snapshot info")
	volSink, err := newSink(logger.Named("sink"), o.fileSinkPath, o.nomadSinkPath)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("setting up unix socket", "path", o.csiSock)
	if err := os.Remove(o.csiSock); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	l, err := net.Listen("unix", o.csiSock)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", o.csiSock, err)
	}

	logger.Info("starting server", "sock", o.csiSock)
	serv := server.New(logger, o.nodeID, volSink)
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

func getOpts(args []string) (opts, error) {
	o := opts{}
	flags := flag.NewFlagSet("csi-madrid", flag.ExitOnError)
	flags.StringVar(&o.csiSock, "csi-endpoint", "", "path to CSI unix socket (required)")
	flags.StringVar(&o.nodeID, "node-id", "", "node identifier")
	flags.StringVar(&o.nomadSinkPath, "sink-nomad-path", "", "base path for nomad variable store (optional)")
	flags.StringVar(&o.fileSinkPath, "sink-file-path", "", "base path for file store (optional)")
	flags.StringVar(&o.logLevel, "log-level", "info", "log level to output")
	if err := flags.Parse(args); err != nil {
		return o, err
	}
	err := o.validate()
	if err != nil {
		flags.Usage()
	}
	return o, err
}

type opts struct {
	csiSock string
	nodeID  string

	nomadSinkPath string
	fileSinkPath  string

	logLevel string
}

func (o opts) validate() error {
	if o.csiSock == "" {
		return errors.New("csi-endpoint is required")
	}
	if o.nomadSinkPath != "" && o.fileSinkPath != "" {
		return errors.New("sink-file-path and sink-nomad-path are mutually exclusive")
	}
	return nil
}

func newSink(log hclog.Logger, filePath, nomadPath string) (sink.Sink, error) {
	if nomadPath != "" {
		log.Debug("using nomad sink", "path", nomadPath)
		conf := api.DefaultConfig()
		client, err := api.NewClient(conf)
		if err != nil {
			return nil, fmt.Errorf("error getting nomad client: %w", err)
		}
		return sink.NewNomadSink(
			log.Named("nomad"),
			client.Variables(), nomadPath,
			conf.Region, conf.Namespace), nil
	}
	if filePath != "" {
		log.Debug("using file sink", "path", filePath)
		return sink.NewFileSink(log.Named("file"), filePath)
	}
	log.Debug("using memory sink")
	return sink.NewMemSink(), nil
}
