package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONResponse is the stable envelope for -o json output.
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

var format = "table"

// SetFormat stores the active output format from the root command.
func SetFormat(value string) {
	format = value
}

// Format returns the active output format.
func Format() string {
	return format
}

// WriteJSON encodes v with indentation.
func WriteJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// WriteSuccessJSON writes a successful JSON payload.
func WriteSuccessJSON(w io.Writer, data interface{}) error {
	return WriteJSON(w, JSONResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// NotImplemented returns a consistent stub error for unfinished commands.
func NotImplemented(feature string) error {
	return fmt.Errorf("%s 尚未实现，将在后续版本提供", feature)
}
