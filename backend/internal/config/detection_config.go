package config

import "time"

type DetectionConfig struct {
	BruteForceThreshold int           // например, 10
	BruteForceWindow    time.Duration // например, 1m
}
