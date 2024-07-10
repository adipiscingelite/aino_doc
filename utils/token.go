package utils

var InvalidTokens = make(map[string]struct{})

func InvalidateToken(token string) {
	InvalidTokens[token] = struct{}{}
}
