package api

import "github.com/rs/zerolog"

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}
