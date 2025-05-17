package cmd_test

import (
	"testing"

	"github.com/religiosa1/tgnotifier/internal/cmd"
	"github.com/stretchr/testify/assert"
)

func TestSend_parseFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"short", []string{"-r", "1,2,3", "-m", "HTML", "-c", "test-conf", "--bot-token", "test-token", "lorem"}},
		{"long", []string{"--recipients", "1,2,3", "--parse-mode", "HTML", "--config", "test-conf", "--bot-token", "test-token", "lorem"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var cmd cmd.Send
			p := newCliParserWithConfig(t, &cmd, mockConfig)
			_, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("error parsing args: %v", err)
			}
			assert.Equal(t, "HTML", cmd.ParseMode)
			assert.Equal(t, []string{"1", "2", "3"}, cmd.Recipients)
			assert.Equal(t, "test-conf", cmd.Config)
			assert.Equal(t, "test-token", cmd.BotToken)
			assert.Equal(t, "lorem", cmd.Message)
		})
	}
}

func TestSend_parseWithDefaultsFromConfig(t *testing.T) {
	var cmd cmd.Send
	p := newCliParserWithConfig(t, &cmd, mockConfig)
	_, err := p.Parse([]string{"ipsum"})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, "", cmd.ParseMode) // no defaults for parseMode, that's for Bot to handle
	assert.Equal(t, mockConfig.Recipients, cmd.Recipients)
	assert.Equal(t, mockConfig.BotToken, cmd.BotToken)
	assert.Equal(t, "ipsum", cmd.Message)
}

// TODO test no config

func TestSend_parseWithEnv(t *testing.T) {
	t.Setenv("BOT_RECIPIENTS", "7,8,9")
	t.Setenv("BOT_TOKEN", "test-token")

	var cmd cmd.Send
	p := newCliParserWithConfig(t, &cmd, mockConfig)
	_, err := p.Parse([]string{"-m", "HTML"})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, "HTML", cmd.ParseMode)
	assert.Equal(t, []string{"7", "8", "9"}, cmd.Recipients)
	assert.Equal(t, "test-token", cmd.BotToken)
}

func TestSend_parsePriorityFlagEnvConfig(t *testing.T) {
	t.Setenv("BOT_RECIPIENTS", "7,8,9")
	t.Setenv("BOT_TOKEN", "test-token")

	var cmd cmd.Send
	p := newCliParserWithConfig(t, &cmd, mockConfig)
	_, err := p.Parse([]string{"-m", "HTML", "-r", "5,6,7"})
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	assert.Equal(t, "HTML", cmd.ParseMode)                   // flag without env
	assert.Equal(t, []string{"5", "6", "7"}, cmd.Recipients) // flag overrides env
	assert.Equal(t, "test-token", cmd.BotToken)              // env without flag
}

// TODO test validation
