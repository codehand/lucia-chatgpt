package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const (
	openAiHost = "https://api.openai.com"
)

var (
	ModeHTML = &tele.SendOptions{
		ParseMode: tele.ModeHTML,
	}
)

// AIBot ...
type AIBot struct {
	bot    *tele.Bot
	apiKey string
	model  string
	client *http.Client
	cache  *bigcache.BigCache
} // @name AIBot

// CreateCompletionRequest ...
type CreateCompletionRequest struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	TopP             int      `json:"top_p"`
	FrequencyPenalty float64  `json:"frequency_penalty"`
	PresencePenalty  float64  `json:"presence_penalty"`
	Stop             []string `json:"stop"`
} // @name CreateCompletionRequest

// CreateCompletionResponse ...
type CreateCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     any    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
} // @name CreateCompletionResponse

// CreateCompletionGPTRequest ...
type CreateCompletionGPTRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             int       `json:"top_p"`
	FrequencyPenalty float64   `json:"frequency_penalty"`
	PresencePenalty  float64   `json:"presence_penalty"`
	Stop             []string  `json:"stop"`
} // @name CreateCompletionGPTRequest

// Message ...
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CreateCompletionError ...
type CreateCompletionError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   any    `json:"param"`
		Code    any    `json:"code"`
	} `json:"error"`
} // @name CreateCompletionError

// NewAIBot ...
func NewAIBot(teleApiKey, openApiKey string) *AIBot {

	pref := tele.Settings{
		Token:  teleApiKey,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	b.Use(middleware.Logger())
	b.Use(middleware.AutoRespond())

	// init cache
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))

	aiBot := &AIBot{
		bot:    b,
		model:  "gpt-3.5-turbo",
		apiKey: openApiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cache,
	}

	// START
	aiBot.bot.Handle("/start", func(c tele.Context) error {
		return c.Send("Hi! I'm <b>ChatGPT</b> bot implemented with GPT-3.5 OpenAI API 🤖\n\n", ModeHTML)
	})

	aiBot.bot.Handle("/clear", func(c tele.Context) error {
		var (
			user   = c.Sender()
			userID = fmt.Sprint(user.ID)
		)
		aiBot.cache.Get(userID)
		return c.Send("Hi! I'm <b>ChatGPT</b> bot implemented with GPT-3.5 OpenAI API 🤖\n\n", ModeHTML)
	})

	aiBot.bot.Handle("/help", func(c tele.Context) error {
		helpStr := `Commands:
		⚡️ /start - Register chat
		⚡️ /clear - Open new conversation
		⚡️ /ai - Suggest more other bot ai
		⚡️ /help - Show help
		`
		return c.Send(helpStr, ModeHTML)
	})

	aiBot.bot.Handle(tele.OnText, func(c tele.Context) error {
		var (
			user   = c.Sender()
			text   = c.Text()
			t      = time.Now()
			msgs   []Message
			userID = fmt.Sprint(user.ID)
		)
		if len(text) == 0 {
			return c.Send("Can I help you, give me some question !", ModeHTML)
		}

		c.Send(tele.Typing)
		cacheMsgs, err := aiBot.cache.Get(userID)
		if err == nil {
			_ = json.Unmarshal(cacheMsgs, &msgs)
		}

		respText, err := aiBot.createCompletion(text, msgs)
		if err != nil {
			respText = "OPS ! Bot busy, please try again"
		}
		msgs = append(msgs, Message{
			Role:    "user",
			Content: text,
		})
		msgs = append(msgs, Message{
			Role:    "assistant",
			Content: respText,
		})

		if len(msgs) > 10 {
			msgs = msgs[2:]
		}
		cmsgs, _ := json.Marshal(msgs)
		aiBot.cache.Set(userID, cmsgs)

		respText += fmt.Sprintf("\n (%.2fs) ", time.Since(t).Seconds())
		_, err = b.Send(user, respText, ModeHTML)
		return err
	})

	return aiBot
}

// Start ...
func (b *AIBot) Start() {
	fmt.Println("Bot running ...")
	b.bot.Start()
}

// createCompletion return result then call openai platform
func (b *AIBot) createCompletion(query string, oldMessages []Message) (string, error) {
	// JSON body
	var (
		payload []byte
		err     error
	)

	if b.model == "gpt-3.5-turbo" {
		finalMessage := oldMessages
		finalMessage = append(finalMessage, Message{
			Role:    "user",
			Content: query,
		})
		payload, err = json.Marshal(&CreateCompletionGPTRequest{
			Model:            b.model,
			Messages:         finalMessage,
			Temperature:      0.7,
			MaxTokens:        1000,
			TopP:             1,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stop:             []string{" Human:", " Ai:"},
		})
		if err != nil {
			return "", err
		}

	} else {
		payload, err = json.Marshal(&CreateCompletionRequest{
			Model:            "text-davinci-003",
			Prompt:           query,
			Temperature:      0.7,
			MaxTokens:        1000,
			TopP:             1,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stop:             []string{" Human:", " Ai:"},
		})
		if err != nil {
			return "", err
		}

	}

	// if use model `text-davinci-003`, you should change path to /v1/completions
	// https://platform.openai.com/docs/api-reference/chat/create
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/chat/completions", openAiHost), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", b.apiKey))

	res, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var data CreateCompletionResponse
	err = json.Unmarshal(resBody, &data)
	if err != nil {
		fmt.Printf("Err=%s\n", err.Error())
		return "", err
	}

	choiceTexts := []string{}
	if len(data.Choices) > 0 {
		for _, c := range data.Choices {
			if c.Text != "" {
				choiceTexts = append(choiceTexts, c.Text)
				continue
			}
			choiceTexts = append(choiceTexts, c.Message.Content)
		}
	}

	if len(choiceTexts) > 0 {
		return strings.Join(choiceTexts, "\n\n"), nil
	}

	return "Something went wrong, please retry later", nil
}
