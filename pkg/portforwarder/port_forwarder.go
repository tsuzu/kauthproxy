// Package portforwarder provides port forwarding between local and Kubernetes.
package portforwarder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/google/wire"
	"github.com/int128/kauthproxy/pkg/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var Set = wire.NewSet(
	wire.Struct(new(PortForwarder), "*"),
	wire.Bind(new(Interface), new(*PortForwarder)),
)

//go:generate mockgen -destination mock_portforwarder/mock_portforwarder.go github.com/int128/kauthproxy/pkg/portforwarder Interface

type Interface interface {
	Start(ctx context.Context, eg *errgroup.Group, o Options) error
}

// PortForwarder provides port forwarding from a local port to a pod container port.
type PortForwarder struct {
	Logger logger.Interface
}

// Options represents an option of PortForwarder.
type Options struct {
	Config *rest.Config
	Source Source
	Target Target
}

// Source represents a local source.
type Source struct {
	Port int
}

// Target represents a target pod.
type Target struct {
	Pod           *v1.Pod
	ContainerPort int
}

// Start starts port forwarding in goroutines.
func (pf *PortForwarder) Start(ctx context.Context, eg *errgroup.Group, o Options) error {
	pfURL, err := url.Parse(o.Config.Host + o.Target.Pod.GetSelfLink() + "/portforward")
	if err != nil {
		return xerrors.Errorf("could not build URL for portforward: %w", err)
	}
	rt, upgrader, err := spdy.RoundTripperFor(o.Config)
	if err != nil {
		return xerrors.Errorf("could not create a round tripper: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: rt}, http.MethodPost, pfURL)

	portPair := fmt.Sprintf("%d:%d", o.Source.Port, o.Target.ContainerPort)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{})
	forwarder, err := portforward.New(dialer, []string{portPair}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return xerrors.Errorf("could not create a port forwarder: %w", err)
	}
	eg.Go(func() error {
		<-ctx.Done()
		pf.Logger.V(1).Infof("stopping the port forwarder")
		close(stopChan)
		return nil
	})
	eg.Go(func() error {
		pf.Logger.V(1).Infof("starting a port forwarder for %s", portPair)
		if err := forwarder.ForwardPorts(); err != nil {
			return xerrors.Errorf("could not run the forwarder: %w", err)
		}
		pf.Logger.V(1).Infof("stopped the port forwarder")
		return nil
	})
	return nil
}
