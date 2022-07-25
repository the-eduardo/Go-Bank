package util

type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress strig  `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {

}
