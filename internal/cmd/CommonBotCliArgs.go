package cmd

type CommonBotCliArgs struct {
	Config     string   `env:"BOT_CONFIG_PATH" default:"${config_file}" short:"c" help:"Configuration file name" type:"yamlfile"`
	Recipients []string `env:"BOT_RECIPIENTS" yaml:"recipients" short:"r" required:"" help:"Message recipients, comma separated (defaults to value from config or env)"`
	BotToken   string   `env:"BOT_TOKEN" yaml:"bot_token" required:"" help:"Your bot token as given by botfather (defaults to value from config or env)"`
}

// TODO parse config file write here
