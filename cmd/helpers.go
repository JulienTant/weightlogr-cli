package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

func loadTimezone() (*time.Location, error) {
	tz, err := time.LoadLocation(viper.GetString("timezone"))
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %w", err)
	}
	return tz, nil
}
