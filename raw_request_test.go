package bot

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type clientMock struct {
	requestURI string
}

func (c *clientMock) Do(req *http.Request) (*http.Response, error) {
	c.requestURI = req.URL.RequestURI()
	if c.requestURI == "failed/botXXX/foo" {
		return nil, errors.New(c.requestURI)
	}
	resp := http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
	}
	return &resp, nil
}

func Test_rawRequest_url(t *testing.T) {
	cm := &clientMock{}
	b := &Bot{
		token:  "XXX",
		client: cm,
	}

	err := b.rawRequest(context.Background(), "foo", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.requestURI != "/botXXX/foo" {
		t.Fatalf("unexpected requestURI: %s", cm.requestURI)
	}
}

func Test_rawRequest_url_testEnv(t *testing.T) {
	cm := &clientMock{}
	b := &Bot{
		token:           "XXX",
		client:          cm,
		testEnvironment: true,
	}

	err := b.rawRequest(context.Background(), "foo", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cm.requestURI != "/botXXX/test/foo" {
		t.Fatalf("unexpected requestURI: %s", cm.requestURI)
	}
}

func Test_rawRequest_err_hideToken(t *testing.T) {
	cm := &clientMock{}
	b := &Bot{
		url:    "failed",
		token:  "XXX",
		client: cm,
	}

	err := b.rawRequest(context.Background(), "foo", nil, nil)
	if err == nil {
		t.Fatalf("unexpected nil error")
	}

	if strings.Contains(err.Error(), "XXX") {
		t.Fatalf("unexpected error with token: %s", err.Error())
	}
}
