package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

const maxJSONBodyBytes = 1 << 20 // 1MB

func DecodeJSONBody(r *http.Request, dst any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("content-type header is required")
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.New("invalid content-type")
	}
	if !strings.EqualFold(mediaType, "application/json") {
		return errors.New("content-type must be application/json")
	}

	limitedReader := io.LimitReader(r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(limitedReader)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only one json object")
	}

	return nil
}
