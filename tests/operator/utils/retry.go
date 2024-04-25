package utils

import (
	"time"

	"github.com/avast/retry-go"
)

func WithRetry(utils *TestUtils, f func(utils *TestUtils) error) error {
	return retry.Do(
		func() error {
			return f(utils)
		},
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(100),
		retry.Context(utils.Ctx),
		retry.LastErrorOnly(true),
	)
}
