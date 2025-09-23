package config

type DbConfig struct {
	Path string
}

type OpenAIConfig struct {
	SecretKey string
}

type Config struct {
	Db DbConfig

	Ai OpenAIConfig
}
