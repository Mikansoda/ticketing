package service

import(
	"time"
	"ticketing/helper"
)

func ParseAccessForMiddleware(token string) (*helper.JWTClaims, error) {
	return helper.ParseAccess(token)
}

func AccessBlacklistLookup(token string) (time.Time, bool) {
	exp, ok := accessBlacklist[token]
	return exp, ok
}
