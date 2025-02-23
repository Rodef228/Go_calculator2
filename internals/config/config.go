package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	TIME_ADDITION_MS        time.Duration
	TIME_SUBTRACTION_MS     time.Duration
	TIME_MULTIPLICATIONS_MS time.Duration
	TIME_DIVISIONS_MS       time.Duration
	COMPUTING_POWER         int
}

func Load() Config {
	cfg := Config{}

	add, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	sub, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	mult, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	div, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))

	cfg.TIME_ADDITION_MS = time.Duration(add) * time.Millisecond
	cfg.TIME_SUBTRACTION_MS = time.Duration(sub) * time.Millisecond
	cfg.TIME_MULTIPLICATIONS_MS = time.Duration(mult) * time.Millisecond
	cfg.TIME_DIVISIONS_MS = time.Duration(div) * time.Millisecond
	cfg.COMPUTING_POWER, _ = strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	return cfg
}
