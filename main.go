package main

import (
	"os"

	"github.com/codehand/lucia-chatgpt/bot"
)

func main() {
	b := bot.NewAIBot(os.Getenv("TELE_API_KEY"), os.Getenv("OPENAI_API_KEY"))
	b.Start()
}
