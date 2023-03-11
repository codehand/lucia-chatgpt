# ChatGPT Telegram Bot

ChatGPT is a free online chat platform that allows people from all over the world to communicate with each other. Whether you're looking to meet new friends or want to engage in meaningful conversations, ChatGPT offers a safe and secure environment where you can express yourself and connect with others. With a wide range of chat rooms available, you can find the perfect space to chat with people who share your interests, hobbies, and beliefs. The platform is easy to use, and you can access it from any device connected to the internet. Join ChatGPT today and enjoy chatting with people from all walks of life.

## News
- *8 Mar 2023*: Added voice message recognition with [OpenAI Whisper API](https://openai.com/blog/introducing-chatgpt-and-whisper-apis). Record a voice message and ChatGPT will answer you!
- *2 Mar 2023*: Added support of [ChatGPT API](https://platform.openai.com/docs/guides/chat/introduction). It's enabled by default and can be disabled with `use_chatgpt_api` option in config. Don't forget to **rebuild** you docker image (`--build`).

## Bot commands
- `/start` – Register bot
- `/help` – Show helper
- `/ai` – Suggest other ai bot

## Setup
1. Get your [OpenAI API](https://openai.com/api/) or [API-KEY](https://platform.openai.com/account/api-keys) key

2. Get your Telegram bot token from [@BotFather](https://t.me/BotFather)

3. Set local env `OPENAI_API_KEY` and `TELE_API_KEY`

4. Run command `go run main.go`, else not setup env `OPENAI_API_KEY= TELE_API_KEY= go run main.go`

🔥 And now **run**:

```bash
go build -o app-exe

docker-compose up --build -d

or

docker run -d -e OPENAI_API_KEY='' -e TELE_API_KEY='' capzr/lucia-chatgpt-master

```

## References
1. [*Build ChatGPT from GPT-3*](https://learnprompting.org/docs/applied_prompting/build_chatgpt)