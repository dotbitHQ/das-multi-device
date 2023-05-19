package tool

import (
	"fmt"
	"time"
)

func GetDurationFormatBySecond(sec int64) (formatString string) {
	duration := time.Duration(sec) * time.Second
	d := int(duration.Hours()) / 24
	h := int(duration.Hours()) % 24
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60
	if d > 0 {
		formatString = fmt.Sprintf("%dd ", d)
	}
	if h > 0 {
		formatString += fmt.Sprintf("%dh ", h)
	}
	if m > 0 {
		formatString += fmt.Sprintf("%dm ", m)
	}
	if s > 0 {
		formatString += fmt.Sprintf("%ds ", s)
	}
	return
}
