package chat

import "fmt"

var ErrInvalidRequest = fmt.Errorf("invalid chat request")

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("chat api error: status=%d", e.StatusCode)
	}
	return fmt.Sprintf("chat api error: status=%d body=%s", e.StatusCode, e.Body)
}
