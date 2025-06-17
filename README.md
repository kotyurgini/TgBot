# TgBot with AI
---
This telegram bot allows you to send messages to popular AI models.
## Features
- Send message to chat GPT.
- Dialog system.

Working example: @ginaibot

---
## Start
`go run ./cmd/ --config-path "./configs/local.yaml"`
./configs/local.yaml - path to config file.

## Docker
`docker build -t tgbot .`
`docker run -d --name tgaibot -v /data:/app/data tgbot`
/data - directory where is config file located. **Config file must named 'prod.yaml'**. Logs will also be saved here.

---
## Roadmap
- Add support to image, file and voice message.
- Add other AI models.
- Add image generate.
- Add postgres support.
- Add subscribed for increase limits.