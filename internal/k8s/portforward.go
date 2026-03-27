package k8s

import (
	"fmt"
	"io"
	"net/http"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ActivePortForward represents a running port-forward session.
type ActivePortForward struct {
	Ready <-chan struct{}
	Done  <-chan struct{}
	ErrCh <-chan error
	stop  chan struct{}
}

// Stop signals the port-forward to stop and waits for it to finish.
// It is safe to call multiple times.
func (a *ActivePortForward) Stop() {
	select {
	case <-a.stop:
	default:
		close(a.stop)
	}
	<-a.Done
}

// PortForward starts a native port-forward to the given pod using SPDY.
func PortForward(client *Client, podName, namespace string, localPort, remotePort int) (*ActivePortForward, error) {
	reqURL := client.Typed.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(client.Config)
	if err != nil {
		return nil, fmt.Errorf("creating SPDY round-tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, reqURL)

	stopCh := make(chan struct{})
	readyCh := make(chan struct{})
	doneCh := make(chan struct{})
	errCh := make(chan error, 1)

	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}

	fw, err := portforward.New(dialer, ports, stopCh, readyCh, io.Discard, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("creating port-forwarder: %w", err)
	}

	go func() {
		defer close(doneCh)
		if err := fw.ForwardPorts(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	return &ActivePortForward{
		Ready: readyCh,
		Done:  doneCh,
		ErrCh: errCh,
		stop:  stopCh,
	}, nil
}
