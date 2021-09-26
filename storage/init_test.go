package storage

import "github.com/rs/zerolog"

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}
