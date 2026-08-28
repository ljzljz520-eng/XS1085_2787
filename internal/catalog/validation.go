package catalog

import "strings"

func ValidateMessage(message Message) error {
	if strings.TrimSpace(message.Key) == "" {
		return ErrMissingKey
	}
	if strings.TrimSpace(message.Text) == "" {
		return ErrMissingText
	}
	if len(message.Text) > 300 {
		return ErrTooLong
	}
	return nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

var (
	ErrMissingKey  error = validationError("message key is required")
	ErrMissingText error = validationError("message text is required")
	ErrTooLong     error = validationError("message is too long")
)
