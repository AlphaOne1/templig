package templig

import (
	"fmt"
)

// wrapError wraps an existing error with the provided text.
// It returns nil if `err` is nil, or the original `err` if `text` is empty.
// Otherwise, it returns a new error with the format "text: original_error".
func wrapError(text string, err error) error {
	if err == nil {
		return nil
	}

	if text == "" {
		return err
	}

	return fmt.Errorf("%s: %w", text, err)
}
