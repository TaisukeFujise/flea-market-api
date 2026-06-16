package handler

import (
	"bytes"
	"io"
	"net/http"
)

const maxImageSize = 10 << 20 // 10MB

// sniffImage reads up to 512 bytes to detect the MIME type and returns a new
// reader that prepends those bytes back so the full stream is preserved.
func sniffImage(f io.Reader) (contentType string, r io.Reader, err error) {
	buf := make([]byte, 512)
	n, readErr := io.ReadFull(f, buf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", nil, readErr
	}
	return http.DetectContentType(buf[:n]), io.MultiReader(bytes.NewReader(buf[:n]), f), nil
}
