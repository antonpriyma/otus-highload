package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/clients/http/internal/mock"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"
)

func simpleTestClient(httpClient http.Client) Client {
	return &simpleHTTPClient{
		Client:         httpClient,
		Logger:         log.Null,
		LogParams:      LogParams{},
		ProxyBasicAuth: "",
	}
}

func TestSimpleHTTPClientEraseURLQueryParams(t *testing.T) {
	testURL := utils.MustParseURL("https://videotestapi.ok.ru/fb.do?application_key=CPBLJNJGDIHBABABA&id=256&method=auth.getTokenForAnonym&name=1&phone=1&sig=8342382fd648bc7ae6ca9e3412dd9aec") //nolint:lll

	tests := []struct {
		name         string
		secretParams []string
		want         string
	}{
		{
			name:         "nothing secret",
			secretParams: nil,
			want:         "https://videotestapi.ok.ru/fb.do?application_key=CPBLJNJGDIHBABABA&id=256&method=auth.getTokenForAnonym&name=1&phone=1&sig=8342382fd648bc7ae6ca9e3412dd9aec", //nolint:lll
		}, {
			name:         "something secret",
			secretParams: []string{"application_key", "sig"},
			want:         fmt.Sprintf("https://videotestapi.ok.ru/fb.do?application_key=%s&id=256&method=auth.getTokenForAnonym&name=1&phone=1&sig=%s", secretPlaceholder, secretPlaceholder), //nolint:lll
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(tt *testing.T) {
			c := &simpleHTTPClient{
				Logger: log.Default(),
				LogParams: LogParams{
					SecretURLQueryParams: testCase.secretParams,
				},
			}

			got := c.maskSecretParams(*testURL)
			require.Equal(tt, testCase.want, got.String())
		})
	}
}

func TestSimpleHTTPClientPerformVariousRequests(t *testing.T) {
	var (
		testURL     = "http://mail.ru/some"
		testHeaders = http.Header{
			"header":  []string{"val"},
			"header2": []string{"val2"},
		}
		testBody = []byte("test body")
	)

	type TestCase struct {
		Name string

		RequestBuilder   func(ctrl *gomock.Controller) Request
		CheckRequestFunc func(t *testing.T, req *http.Request)
	}
	testCases := []TestCase{
		{
			Name: "request",
			RequestBuilder: func(ctrl *gomock.Controller) Request {
				req := mock.NewMockRequest(ctrl)
				req.EXPECT().URL().Return(testURL)
				req.EXPECT().Method().Return(http.MethodPost)
				return req
			},
			CheckRequestFunc: func(t *testing.T, req *http.Request) {
				require.Equal(t, testURL, req.URL.String())
				require.Equal(t, http.MethodPost, req.Method)
			},
		}, {
			Name: "request with headers",
			RequestBuilder: func(ctrl *gomock.Controller) Request {
				req := mock.NewMockRequestWithHeaders(ctrl)
				req.EXPECT().URL().Return(testURL)
				req.EXPECT().Method().Return(http.MethodPost)
				req.EXPECT().Headers().Return(testHeaders)
				return req
			},
			CheckRequestFunc: func(t *testing.T, req *http.Request) {
				require.Equal(t, testURL, req.URL.String())
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, testHeaders, req.Header)
			},
		}, {
			Name: "request with body",
			RequestBuilder: func(ctrl *gomock.Controller) Request {
				req := mock.NewMockRequestWithBody(ctrl)
				req.EXPECT().URL().Return(testURL)
				req.EXPECT().Method().Return(http.MethodPost)
				req.EXPECT().Body().Return(testBody, nil)
				return req
			},
			CheckRequestFunc: func(t *testing.T, req *http.Request) {
				require.Equal(t, testURL, req.URL.String())
				require.Equal(t, http.MethodPost, req.Method)

				body, err := ioutil.ReadAll(req.Body)
				require.NoError(t, err)
				require.Equal(t, testBody, body)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(tt *testing.T) {
			ctrl := gomock.NewController(tt)
			defer ctrl.Finish()

			req := testCase.RequestBuilder(ctrl)

			resp := mock.NewMockResponse(ctrl)
			resp.EXPECT().ReadFrom(gomock.Any()).Return(nil)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				testCase.CheckRequestFunc(t, r)
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := simpleTestClient(*srv.Client())

			err := c.PerformRequest(context.Background(), req, resp)
			require.NoError(tt, err)
		})
	}
}

func TestSimpleHTTPClientPerformRequestCtxCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req := mock.NewMockRequest(ctrl)
	req.EXPECT().URL().AnyTimes().Return("http://mail.ru")
	req.EXPECT().Method().AnyTimes().Return(http.MethodGet)

	resp := mock.NewMockResponse(ctrl)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := simpleTestClient(*srv.Client())

	err := c.PerformRequest(ctx, req, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), context.Canceled.Error())
}

func TestSimpleHTTPClientPerformRequestResponseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("some error happened")

	req := mock.NewMockRequest(ctrl)
	req.EXPECT().URL().AnyTimes().Return("http://mail.ru")
	req.EXPECT().Method().AnyTimes().Return(http.MethodGet)

	resp := mock.NewMockResponse(ctrl)
	resp.EXPECT().ReadFrom(gomock.AssignableToTypeOf(&http.Response{})).Return(expectedErr)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := simpleTestClient(*srv.Client())

	err := c.PerformRequest(context.Background(), req, resp)
	require.Error(t, err)
	require.True(t, errors.Is(err, expectedErr))
}
