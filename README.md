# Stand Claimer Bot

Telegram bot for managing dev environments within your team.

## Features

- Automatic notifications for stands held > n hours
- Interactive buttons for claiming/releasing stands
- Stand usage duration tracking
- User management through chat members
- Feature state checking
## Commands

- `/list` - Show all stands with their status and ownership duration
- `/claim` - Claim available stand via interactive buttons
- `/release` - Release your stand
- `/ping` - Ping specific stand owner
- `/ping_all` - Ping all users with busy stands
- `/features_state` - Show current state of the features

## Quick Start

1. Clone the repository
2. Create `.env` file with command ```touch .env``` and configure as it shown in example
3. Create `config.yaml` with command ```mkdir -p ./config && touch ./config/config.yaml``` and configure the following : 
```yaml
    postgres:
      dsn:
        host: db
        port: 5432
        user: postgres
        password: postgres
        db_name: stands
        sslmode: disable
      max_idle_conns: 5
      max_open_conns: 5
      use_seed: true

    bot:
      token: tokenFromENV
      verbose: true

    gitlab:
      token: tokenFromEnv
      url: https://gitlab.com
      project_id: 12345678
      group_id: 00123
```
4. Configure fixtures to preseed your stands by name in stands table. See fixtures/stands.yaml for reference.
5. Run with docker:
```bash
docker-compose up -d --build
```
6. Bot Setup:
- Create bot via [@BotFather](https://t.me/botfather)
- Add bot to team chat
- Start using commands
- Bot listens to UserJoin and UserLeft events to either set or delete users from database
``` NOTE: automatic notifications will start right after bot received any of commands ```

