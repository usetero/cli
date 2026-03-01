package usecase

type StreamErrorMapper interface {
	Classify(err error) string
	UserFacing(err error) string
}

func ClassifyStreamError(mapper StreamErrorMapper, err error) string {
	if mapper == nil {
		return "unknown"
	}
	return mapper.Classify(err)
}

func UserFacingStreamError(mapper StreamErrorMapper, err error) string {
	if mapper == nil {
		return "Chat stream failed. Please try again."
	}
	return mapper.UserFacing(err)
}
