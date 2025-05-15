package jsonx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
)

var (
	ErrInvalidJSON          = errors.New("invalid JSON format")
	ErrInvalidTarget        = errors.New("decode target must be a non-nil pointer")
	ErrNoContent            = errors.New("no content to decode")
	ErrUnsupportedMediaType = errors.New("unsupported media type, expected application/json")
)

type Options struct {
	SuccessStatus int
	ErrorStatus   int

	ContentType        string
	AllowEmpty         bool
	EnforceContentType bool
	Headers            map[string]string

	IndentResponse bool
	EscapeHTML     bool
}

func DefaultOptions() Options {
	return Options{
		SuccessStatus:      http.StatusOK,
		ErrorStatus:        http.StatusBadRequest,
		ContentType:        "application/json",
		AllowEmpty:         true,
		EnforceContentType: true,
		Headers:            map[string]string{},
		IndentResponse:     false,
		EscapeHTML:         false,
	}
}

// ErrorDetail represents a structured error with code and message
type ErrorDetail struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Response struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
	Error   any  `json:"error"`
	Meta    any  `json:"meta"`
}

func mergeOptions(defaults Options, customs ...Options) Options {
	result := defaults

	if len(customs) == 0 {
		return result
	}

	custom := customs[0]

	if custom.SuccessStatus != 0 {
		result.SuccessStatus = custom.SuccessStatus
	}

	if custom.ErrorStatus != 0 {
		result.ErrorStatus = custom.ErrorStatus
	}

	if custom.ContentType != "" {
		result.ContentType = custom.ContentType
	}

	result.AllowEmpty = custom.AllowEmpty
	result.EnforceContentType = custom.EnforceContentType
	result.IndentResponse = custom.IndentResponse
	result.EscapeHTML = custom.EscapeHTML

	if custom.Headers != nil {
		result.Headers = custom.Headers
	}

	if result.SuccessStatus <= 0 {
		result.SuccessStatus = http.StatusOK
	}

	if result.ErrorStatus <= 0 {
		result.ErrorStatus = http.StatusBadRequest
	}

	return result
}

// DecodeJSON decodes JSON from an io.Reader into the provided target
func DecodeJSON(r io.Reader, target any) error {
	if target == nil || reflect.ValueOf(target).Kind() != reflect.Ptr || reflect.ValueOf(target).IsNil() {
		return ErrInvalidTarget
	}

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(target); err != nil {
		if err == io.EOF {
			return ErrNoContent
		}
		return ErrInvalidJSON
	}

	return nil
}

// DecodeJSONFromRequest decodes JSON from an HTTP request body
func DecodeJSONFromRequest(r *http.Request, target any, opts ...Options) error {
	opt := mergeOptions(DefaultOptions(), opts...)

	if opt.EnforceContentType {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(strings.ToLower(contentType), "application/json") {
			return ErrUnsupportedMediaType
		}
	}

	return DecodeJSON(r.Body, target)
}

// EncodeJSON encodes data to JSON and writes it to the provided writer
func EncodeJSON(w io.Writer, data any, opts ...Options) error {
	opt := mergeOptions(DefaultOptions(), opts...)

	if !opt.AllowEmpty && (data == nil || (reflect.ValueOf(data).Kind() == reflect.Ptr && reflect.ValueOf(data).IsNil())) {
		return ErrNoContent
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(opt.EscapeHTML)
	if opt.IndentResponse {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(data)
}

// RespondWithJSON writes a JSON response with appropriate headers
func RespondWithJSON(w http.ResponseWriter, data any, opts ...Options) error {
	opt := mergeOptions(DefaultOptions(), opts...)

	w.Header().Set("Content-Type", opt.ContentType)
	for k, v := range opt.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(opt.SuccessStatus)

	return EncodeJSON(w, data, opt)
}

// RespondWithError writes a JSON error response
func RespondWithError(w http.ResponseWriter, err any, opts ...Options) error {
	opt := mergeOptions(DefaultOptions(), opts...)

	resp := Response{
		Success: false,
	}

	switch e := err.(type) {
	case ErrorDetail:
		resp.Error = e
	case *ErrorDetail:
		if e != nil {
			resp.Error = *e
		} else {
			resp.Error = "unknown error"
		}
	case error:
		resp.Error = e.Error()
	case string:
		resp.Error = e
	case map[string]any:
		resp.Error = e
	default:
		// Assumes it is JSON serializable
		if err != nil {
			resp.Error = err
		} else {
			resp.Error = "unknown error"
		}
	}

	w.Header().Set("Content-Type", opt.ContentType)
	for k, v := range opt.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(opt.ErrorStatus)

	return EncodeJSON(w, resp, opt)
}

// RespondWithSuccess writes a standardized success response
func RespondWithSuccess(w http.ResponseWriter, data any, meta any, opts ...Options) error {
	opt := mergeOptions(DefaultOptions(), opts...)

	resp := Response{
		Success: true,
		Data:    data,
	}

	if meta != nil {
		resp.Meta = meta
	}

	w.Header().Set("Content-Type", opt.ContentType)
	for k, v := range opt.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(opt.SuccessStatus)

	return EncodeJSON(w, resp, opt)
}

// Send is a shorthand for RespondWithJSON with default options
func Send(w http.ResponseWriter, data any) error {
	return RespondWithJSON(w, data)
}

// SendError is a shorthand for RespondWithError with default options
func SendError(w http.ResponseWriter, err error) error {
	return RespondWithError(w, err)
}

// SendSuccess is a shorthand for RespondWithSuccess with default options
func SendSuccess(w http.ResponseWriter, data any) error {
	return RespondWithSuccess(w, data, nil)
}
