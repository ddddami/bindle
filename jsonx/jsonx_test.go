package jsonx

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		json    string
		target  any
		wantErr error
	}{
		{
			name:    "valid JSON",
			json:    `{"name":"test","value":42}`,
			target:  &TestStruct{},
			wantErr: nil,
		},
		{
			name:    "invalid JSON",
			json:    `{"name":"test","value":}`,
			target:  &TestStruct{},
			wantErr: ErrInvalidJSON,
		},
		{
			name:    "empty input",
			json:    "",
			target:  &TestStruct{},
			wantErr: ErrNoContent,
		},
		{
			name:    "nil target",
			json:    `{"name":"test","value":42}`,
			target:  nil,
			wantErr: ErrInvalidTarget,
		},
		{
			name:    "non-pointer target",
			json:    `{"name":"test","value":42}`,
			target:  TestStruct{},
			wantErr: ErrInvalidTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.json)
			err := DecodeJSON(reader, tt.target)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DecodeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				ts := tt.target.(*TestStruct)
				if ts.Name != "test" || ts.Value != 42 {
					t.Errorf("DecodeJSON() didn't properly decode. Got %+v", ts)
				}
			}
		})
	}

	err := DecodeJSON(nil, &TestStruct{})
	if err != ErrNoContent {
		t.Errorf("DecodeJSON() with nil reader should return ErrNoContent, got %v", err)
	}
}

func TestEncodeJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name       string
		data       any
		opts       Options
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "basic struct",
			data:       TestStruct{Name: "test", Value: 42},
			opts:       DefaultOptions(),
			wantPrefix: `{"name":"test","value":42}`,
			wantErr:    false,
		},
		{
			name: "indented response",
			data: TestStruct{Name: "test", Value: 42},
			opts: Options{
				IndentResponse: true,
				EscapeHTML:     false,
			},
			wantPrefix: "{\n  \"name\": \"test\",",
			wantErr:    false,
		},
		{
			name: "nil data with allowEmpty=false",
			data: nil,
			opts: Options{
				AllowEmpty: false,
			},
			wantPrefix: "",
			wantErr:    true,
		},
		{
			name: "nil data with allowEmpty=true",
			data: nil,
			opts: Options{
				AllowEmpty: true,
			},
			wantPrefix: "null",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := EncodeJSON(buf, tt.data, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.HasPrefix(buf.String(), tt.wantPrefix) {
				t.Errorf("EncodeJSON() = %v, want prefix %v", buf.String(), tt.wantPrefix)
			}
		})
	}
}

func TestDecodeJSONFromRequest(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		enforceType bool
		wantErr     bool
	}{
		{
			name:        "valid request",
			body:        `{"name":"test","value":42}`,
			contentType: "application/json",
			enforceType: true,
			wantErr:     false,
		},
		{
			name:        "invalid content type",
			body:        `{"name":"test","value":42}`,
			contentType: "text/plain",
			enforceType: true,
			wantErr:     true,
		},
		{
			name:        "invalid json",
			body:        `{"name":"test","value":}`,
			contentType: "application/json",
			enforceType: true,
			wantErr:     true,
		},
		{
			name:        "ignore content type",
			body:        `{"name":"test","value":42}`,
			contentType: "text/plain",
			enforceType: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			target := &TestStruct{}
			err := DecodeJSONFromRequest(req, target, Options{EnforceContentType: tt.enforceType})

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSONFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if target.Name != "test" || target.Value != 42 {
					t.Errorf("DecodeJSONFromRequest() didn't properly decode. Got %+v", target)
				}
			}
		})
	}
}

func TestRespondFunctions(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(http.ResponseWriter) error
		wantStatus int
		wantBody   string
	}{
		{
			name: "Send",
			fn: func(w http.ResponseWriter) error {
				return Send(w, map[string]any{"key": "value"})
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"key":"value"}`,
		},
		{
			name: "SendError",
			fn: func(w http.ResponseWriter) error {
				return SendError(w, errors.New("test error"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"success":false,"data":null,"error":"test error","meta":null}`,
		},
		{
			name: "SendSuccess",
			fn: func(w http.ResponseWriter) error {
				return SendSuccess(w, map[string]any{"key": "value"})
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"success":true,"data":{"key":"value"},"error":null,"meta":null}`,
		},
		{
			name: "RespondWithJSON custom status",
			fn: func(w http.ResponseWriter) error {
				return RespondWithJSON(w, "test data", Options{SuccessStatus: http.StatusCreated})
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"test data"`,
		},
		{
			name: "RespondWithError custom status",
			fn: func(w http.ResponseWriter) error {
				return RespondWithError(w, errors.New("not found"), Options{ErrorStatus: http.StatusNotFound})
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"success":false,"data":null,"error":"not found","meta":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			err := tt.fn(rr)
			if err != nil {
				t.Errorf("Function returned error: %v", err)
				return
			}

			if rr.Code != tt.wantStatus {
				t.Errorf("Status code = %d, want %d", rr.Code, tt.wantStatus)
			}

			if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ctype)
			}

			body := strings.TrimSpace(rr.Body.String())
			if !strings.HasPrefix(body, strings.TrimSpace(tt.wantBody)) {
				t.Errorf("Body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
