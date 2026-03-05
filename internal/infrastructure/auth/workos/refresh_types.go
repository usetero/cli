package workos

// RefreshResult is returned by refresh token flow.
type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}
