package logger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBodyLogWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "should write string content",
			input:    []byte("test content"),
			expected: "test content",
		},
		{
			name:     "should write json content",
			input:    []byte(`{"key":"value"}`),
			expected: `{"key":"value"}`,
		},
		{
			name:     "should write empty content",
			input:    []byte(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			w := newTestResponseWriter()
			blw := &BodyLogWriter{
				ResponseWriter: w,
				Body:           bytes.NewBufferString(""),
			}

			// Act
			n, err := blw.Write(tt.input)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, len(tt.input), n)
			assert.Equal(t, tt.expected, blw.Body.String())
			assert.Equal(t, tt.expected, w.Body.String())
		})
	}
}

func TestBodyLogWriter_Creation(t *testing.T) {
	t.Run("should create new body log writer", func(t *testing.T) {
		// Arrange
		w := newTestResponseWriter()

		// Act
		blw := &BodyLogWriter{
			ResponseWriter: w,
			Body:           bytes.NewBufferString(""),
		}

		// Assert
		assert.NotNil(t, blw)
		assert.NotNil(t, blw.Body)
		assert.NotNil(t, blw.ResponseWriter)
		assert.Empty(t, blw.Body.String())
	})
}
