package config

type Config[T any] interface {
	WithDefaults() Config[T]
	WithJSON() Config[T]
	WithFlags() Config[T]
	WithEnv() Config[T]
	Build() (*T, error)
}
