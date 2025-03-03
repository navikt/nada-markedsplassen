package errs

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi/logger"
)

func Test_httpErrorStatusCode(t *testing.T) {
	type args struct {
		k Kind
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Exist", args{k: Exist}, http.StatusBadRequest},
		{"NotExist", args{k: NotExist}, http.StatusNotFound},
		{"Invalid", args{k: Invalid}, http.StatusBadRequest},
		{"Private", args{k: Private}, http.StatusBadRequest},
		{"BrokenLink", args{k: BrokenLink}, http.StatusBadRequest},
		{"Validation", args{k: Validation}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"Other", args{k: Other}, http.StatusInternalServerError},
		{"IO", args{k: IO}, http.StatusInternalServerError},
		{"Internal", args{k: Internal}, http.StatusInternalServerError},
		{"Database", args{k: Database}, http.StatusInternalServerError},
		{"Unanticipated", args{k: Unanticipated}, http.StatusInternalServerError},
		{"Default", args{k: 99}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpErrorStatusCode(tt.args.k); got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorResponse_StatusCode(t *testing.T) {
	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	l := logger.New(os.Stdout, zerolog.DebugLevel, false)

	unauthenticatedErr := E(Unauthenticated, "some error from Google")
	unauthorizedErr := E(Unauthorized, "some authorization error")

	tests := []struct {
		name string
		args args
		want int
	}{
		{"nil error", args{httptest.NewRecorder(), l, nil}, http.StatusInternalServerError},
		{"empty *Error", args{httptest.NewRecorder(), l, &Error{}}, http.StatusInternalServerError},
		{"unauthenticated", args{httptest.NewRecorder(), l, unauthenticatedErr}, http.StatusUnauthorized},
		{"unauthorized", args{httptest.NewRecorder(), l, unauthorizedErr}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPErrorResponse(tt.args.w, l, tt.args.err, "")
			if got := tt.args.w.Result().StatusCode; got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorResponse_Body(t *testing.T) {
	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	var b bytes.Buffer
	lgr := logger.New(&b, zerolog.DebugLevel, false)

	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty Error", args{httptest.NewRecorder(), lgr, &Error{}}, `{"error":{"kind":"unanticipated_error","code":"Unanticipated","message":"errors: other_error"}}`},
		{"unauthenticated", args{httptest.NewRecorder(), lgr, E(Unauthenticated, "some error from Google")}, `{"error":{"kind":"unauthenticated_request","statusCode":401,"message":"some error from Google","requestId":"abc"}}`},
		{"unauthorized", args{httptest.NewRecorder(), lgr, E(Unauthorized, "some authorization error")}, `{"error":{"kind":"unauthorized_request","statusCode":403,"message":"some authorization error","requestId":"abc"}}`},
		{"normal", args{httptest.NewRecorder(), lgr, E(Exist, Parameter("some_param"), Code("some_code"), errors.New("some error"))}, `{"error":{"kind":"item_exists","statusCode":400,"code":"some_code","param":"some_param","message":"some error","requestId":"abc"}}`},
		{"not via E", args{httptest.NewRecorder(), lgr, errors.New("some error")}, `{"error":{"kind":"unanticipated_error","code":"Unanticipated","message":"some error"}}`},
		{"nil error", args{httptest.NewRecorder(), lgr, nil}, `{"error":{"kind":"unanticipated_error","code":"Unanticipated","message":"Unexpected error - contact support"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPErrorResponse(tt.args.w, lgr, tt.args.err, "abc")
			if got := strings.TrimSpace(tt.args.w.Body.String()); got != tt.want {
				t.Errorf("httpErrorResponseBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
