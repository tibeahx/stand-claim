# Stand Claimer Bot

Telegram bot for managing dev environments within your team.

## Commands

- `/list` - Show all stands with their status and ownership duration
- `/claim` - Claim available stand via interactive buttons
- `/release` - Release your stand
- `/ping` - Ping specific stand owner
- `/ping_all` - Ping all users with busy stands

## Features

- Automatic notifications for stands held > n hours
- Interactive buttons for claiming/releasing stands
- Stand usage duration tracking
- User management through chat members


## Quick Start

1. Clone the repository
2. Create `.env` file: ```BOT_TOKEN=your_telegram_bot_token ```
3. Configure in `config/config.yaml`: 
```yaml
    postgres:
        dsn:
        host: db
        port: 5432
        user: postgres
        password: postgres
        db_name: stands
        sslmode: disable
        use_seed: true
    bot:
        stands: [dev1, dev2, staging] # your enviroments environments 
```
4. Configure fixtures to seed base data in Postgres
5. Run with docker:
```bash
docker-compose up -d --build
```
6. Bot Setup:
- Create bot via [@BotFather](https://t.me/botfather)
- Add bot to team chat
- Grant admin rights
- Start using commands
``` NOTE: automatic notifications will start right after bot received any of commands ```

