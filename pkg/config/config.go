package config

// example) min: 100m
type minmax map[string]string

// example) cpu.min: 100m
type resources map[string]minmax

type Configuration struct {
	Requests resources
	Limits   resources
}
