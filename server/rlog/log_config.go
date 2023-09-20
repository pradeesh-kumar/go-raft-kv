package rlog

type LogConfig struct {
	MaxStoreBytes uint64 `yaml:"maxStoreBytes"`
	MaxIndexBytes uint64 `yaml:"maxIndexBytes"`
	InitialOffset uint64 `yaml:"initialOffset"`
}
