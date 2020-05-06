package ghapps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/golang/mock/gomock"
	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
)

func TestCron(t *testing.T) {
	cfg := stackdriverlog.NewConfig("")
	cfg.ContextLogOut, cfg.RequestLogOut = ioutil.Discard, ioutil.Discard
	mw := logging.WithLogger(cfg)
	now := time.Now().Truncate(time.Second)

	cases := []struct {
		name         string
		reqBody      interface{}
		statusCode   int
		buildUsecase func(ctrl *gomock.Controller) usecase.Usecase
	}{
		{
			name: "ok",
			reqBody: &PubSubPayload{
				Subscription: "0xdeadbeaf",
				Message: &PubSubMessage{
					Data:        json.RawMessage("{}"),
					ID:          "xxx",
					PublishTime: PublishTime(now),
				},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().UpdateChanceTime(gomock.Any(), gomock.Any(), eqTime(now)).AnyTimes()
				return uc
			},
		},
		{
			name:       "invalid",
			reqBody:    nil,
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
		{
			name: "empty message",
			reqBody: &PubSubPayload{
				Subscription: "0xdeadbead",
			},
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(c.reqBody)

			w := &Web{
				usecase: c.buildUsecase(ctrl),
			}
			srv := httptest.NewServer(mw(w.handleCron()))
			defer srv.Close()

			resp, err := http.Post(srv.URL, "application/json", strings.NewReader(buf.String()))
			if err != nil {
				t.Error(err)
			}
			t.Logf("resp=%#v", resp)
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
		})
	}
}

type timeMatcher struct {
	expected time.Time
}

func (m timeMatcher) Matches(x interface{}) bool {
	t, ok := x.(time.Time)
	if !ok {
		return false
	}
	return t.Equal(m.expected)
}

func (m timeMatcher) String() string {
	return fmt.Sprintf("is equal to %s (%#v)", m.expected, m.expected)
}

func eqTime(expected time.Time) gomock.Matcher {
	return gomock.GotFormatterAdapter(
		gomock.GotFormatterFunc(func(i interface{}) string {
			return fmt.Sprintf("%s (%#v)", i, i)
		}),
		timeMatcher{expected},
	)
}
