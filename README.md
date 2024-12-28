# Stand Claimer Bot

Telegram bot for managing dev environments within your team.


## Features

- Automatic notifications for stands held > n hours
- Interactive buttons for claiming/releasing stands
- Stand usage duration tracking
- User management through chat members

## Commands

- `/list` - Show all stands with their status and ownership duration
- `/claim` - Claim available stand via interactive buttons
- `/release` - Release your stand
- `/ping` - Ping specific stand owner
- `/ping_all` - Ping all users with busy stands

## Usage and deployment 

1. Set .env like: BOT_TOKEN=123456
2. Set your values in config/config.yaml
3. Build and run the image via docker compose
4. Add your bot instance to your team chat
5. Forget about fighting for free enviroment
