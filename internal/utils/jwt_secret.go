package utils

import (
	"os"
)

var (
	AccessTokenSecret  = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	RefreshTokenSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
)
