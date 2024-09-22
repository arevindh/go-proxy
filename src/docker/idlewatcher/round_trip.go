package idlewatcher

import (
	"context"
	"net/http"
	"time"
)

type (
	roundTripper struct {
		patched roundTripFunc
	}
	roundTripFunc func(*http.Request) (*http.Response, error)
)

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.patched(req)
}

func (w *watcher) roundTrip(origRoundTrip roundTripFunc, req *http.Request) (*http.Response, error) {
	// target site is ready, passthrough
	if w.ready.Load() {
		return origRoundTrip(req)
	}

	// wake the container
	w.wakeCh <- struct{}{}

	// initial request
	targetUrl := req.Header.Get(headerGoProxyTargetURL)
	if targetUrl == "" {
		return w.makeResp(
			"%s is starting... Please wait",
			w.ContainerName,
		)
	}

	w.l.Debug("serving event")

	// stream request
	rtDone := make(chan *http.Response, 1)
	ctx, cancel := context.WithTimeout(req.Context(), w.WakeTimeout)
	defer cancel()

	// loop original round trip until success in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.ctx.Done():
				return
			default:
				resp, err := origRoundTrip(req)
				if err == nil {
					w.ready.Store(true)
					rtDone <- resp
					return
				}
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()

	for {
		select {
		case resp := <-rtDone:
			return w.makeSuccResp(targetUrl, resp)
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return w.makeErrResp("Timed out waiting for %s to fully wake", w.ContainerName)
			}
			return w.makeErrResp("idlewatcher has stopped\n%s", w.ctx.Err().Error())
		case <-w.ctx.Done():
			return w.makeErrResp("idlewatcher has stopped\n%s", w.ctx.Err().Error())
		}
	}
}
