package common

import (
	"fmt"
	"strings"
	"time"
)

func TimeNowUTCDay() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func WrapErrors(method string, errors ...error) error {
	res := "%s: " + strings.Repeat("%w: ", len(errors)-1) + "%w"
	args := make([]any, 0, len(errors)+1)
	args = append(args, method)
	for _, err := range errors {
		args = append(args, err)
	}
	return fmt.Errorf(res, args...)
}
