package throttler

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fields struct {
	environment string
	Throttler   *Throttler
}
type args struct {
	attempts int
	token    string
	req      *http.Request
	w        *httptest.ResponseRecorder
}
type response struct {
	code int
	body []byte
	err  error
}

var okResponse = response{
	code: http.StatusOK,
	body: []byte(`[]`),
	err:  nil,
}
var tooManyReq = response{
	code: http.StatusTooManyRequests,
	body: []byte(`Too many requests`),
	err:  nil,
}
var testTokenCache = tokensCache{}

var nextHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`[]`))
})

func TestService_CheckLimitsMiddlware(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		args   args
		res    response
	}{
		{
			name: "check successful request",
			fields: fields{
				environment: "test",
				Throttler: &Throttler{
					N:     1,
					M:     100,
					cache: testTokenCache,
				},
			},
			args: args{
				attempts: 1,
				token:    "ok_token_once",
				req:      httptest.NewRequest("GET", "/v1/users", nil),
				w:        httptest.NewRecorder(),
			},
			res: okResponse,
		},
		{
			name: "check failed for N=0 requests",
			fields: fields{
				environment: "test",
				Throttler: &Throttler{
					N:     0,
					M:     100,
					cache: testTokenCache,
				},
			},
			args: args{
				attempts: 1,
				token:    "rate_limited_token",
				req:      httptest.NewRequest("GET", "/v1/users", nil),
				w:        httptest.NewRecorder(),
			},
			res: tooManyReq,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				environment: tt.fields.environment,
				Throttler:   tt.fields.Throttler,
			}
			ctx := tt.args.req.Context()
			ctx = context.WithValue(ctx, ctxToken, tt.args.token)
			tt.args.req = tt.args.req.WithContext(ctx)

			handler := s.CheckLimitsMiddlware(nextHandler)
			handler.ServeHTTP(tt.args.w, tt.args.req)

			res := tt.args.w.Result()
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)

			if got := res.StatusCode; got != tt.res.code {
				t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res.code)
			}
			if got := err; got != tt.res.err {
				t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res.err)
			}
			if got := string(body); !strings.Contains(got, string(tt.res.body)) {
				t.Errorf("CheckLimitsMiddlware got = %s, want %s", got, tt.res.body)
			}
		})
	}
}

func TestService_CheckLimitsMiddlware_WithMultipleRequests(t *testing.T) {

	tests := []struct {
		name   string
		fields fields
		args   args
		res    []response
	}{
		{
			name: "5 requests at a rate of 1 requests per 200ms",
			fields: fields{
				environment: "test",
				Throttler: &Throttler{
					N:     1,
					M:     50,
					cache: testTokenCache,
				},
			},
			args: args{
				attempts: 5,
				token:    "3_attempts_token",
				req:      httptest.NewRequest("GET", "/v1/users", nil),
				w:        httptest.NewRecorder(),
			},
			res: []response{
				okResponse,
				tooManyReq,
				okResponse,
				tooManyReq,
				okResponse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				environment: tt.fields.environment,
				Throttler:   tt.fields.Throttler,
			}
			for i := 0; i < tt.args.attempts; i++ {
				if i > 0 {
					tt.args.w = httptest.NewRecorder()
				}
				ctx := tt.args.req.Context()
				ctx = context.WithValue(ctx, ctxToken, tt.args.token)
				tt.args.req = tt.args.req.WithContext(ctx)

				handler := s.CheckLimitsMiddlware(nextHandler)
				handler.ServeHTTP(tt.args.w, tt.args.req)

				res := tt.args.w.Result()
				defer res.Body.Close()

				body, err := ioutil.ReadAll(res.Body)

				if got := res.StatusCode; got != tt.res[i].code {
					t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res[i].code)
				}
				if got := err; got != tt.res[i].err {
					t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res[i].err)
				}
				if got := string(body); !strings.Contains(got, string(tt.res[i].body)) {
					t.Errorf("CheckLimitsMiddlware got = %s, want %s", got, tt.res[i].body)
				}
				if res.StatusCode == http.StatusTooManyRequests {
					time.Sleep(time.Duration(tt.fields.Throttler.M) * time.Millisecond)
				}
			}
		})
	}
}
func TestService_ValidateAccessToken(t *testing.T) {
	tests := []struct {
		name   string
		fields fields
		args   args
		res    response
	}{
		{
			name: "check successful request with valid token",
			fields: fields{
				environment: "test",
			},
			args: args{
				token: "ok_token_once",
				req:   httptest.NewRequest("GET", "/v1/users", nil),
				w:     httptest.NewRecorder(),
			},
			res: okResponse,
		},
		{
			name: "Failed on empty token",
			fields: fields{
				environment: "test",
			},
			args: args{
				token: "",
				req:   httptest.NewRequest("GET", "/v1/users", nil),
				w:     httptest.NewRecorder(),
			},
			res: response{
				code: http.StatusUnauthorized,
				body: []byte(`Missing access token`),
				err:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				environment: tt.fields.environment,
			}

			tt.args.req.Header.Set("Authorization", "Bearer "+tt.args.token)

			handler := s.ValidateAccessToken(nextHandler)
			handler.ServeHTTP(tt.args.w, tt.args.req)

			res := tt.args.w.Result()
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)

			if got := res.StatusCode; got != tt.res.code {
				t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res.code)
			}
			if got := err; got != tt.res.err {
				t.Errorf("CheckLimitsMiddlware got = %v, want %v", got, tt.res.err)
			}
			if got := string(body); !strings.Contains(got, string(tt.res.body)) {
				t.Errorf("CheckLimitsMiddlware got = %s, want %s", got, tt.res.body)
			}
		})
	}
}
