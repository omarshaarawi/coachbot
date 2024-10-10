# CoachBot

CoachBot is a Telegram bot designed to provide fantasy football league updates and information for ESPN fantasy leagues.

## Features and Commands

- `/scores`: Get current scores
- `/standings`: View league standings
- `/whohas <player>`: Check which team has a specific player
- `/monitor`: Monitor players with injury status
- `/finalscore`: Get final score reports
- `/mondaynight`: View close games for Monday night
- `/matchup`: See matchups for the current week
- `/start`: Welcome message
- `/help`: List available commands

## Requirements

- Go 1.23 or higher
- Telegram Bot Token
- ESPN Fantasy League credentials

## Configuration

The following environment variables are required:

- `TELEGRAM_TOKEN`: Your Telegram Bot token
- `CHAT_ID`: The Telegram chat ID where the bot will send messages
- `YEAR`: The current NFL season year
- `LEAGUE_ID`: Your ESPN Fantasy Football league ID
- `SWID`: Your ESPN SWID
- `ESPN_S2`: Your ESPN S2 cookie value

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/omarshaarawi/coachbot.git
   cd coachbot
   ```

2. Build the application:
   ```
   make build
   ```

## Usage

To run the bot locally:

```
make run
```

## Scheduler

CoachBot includes a scheduler that automatically sends updates at specific times:

- Monday, Tuesday, Friday at 7:30 CDT: Scoreboard update
- Monday at 17:30 CDT: Close scores for Monday night games
- Tuesday at 7:30 CDT: Weekly trophies report
- Wednesday at 7:30 CDT: Current standings
- Thursday at 18:30 CDT: Matchups for the week
- Sunday at 7:30 CDT: Players to monitor report
- Sunday at 15:00 and 19:00 CDT: Scoreboard updates

The scheduler is configured in `internal/scheduler/scheduler.go`.

## Deployment

The project uses Kamal for deployment. To deploy:

1. Ensure you have Kamal installed and configured.
2. Update the `config/deploy.yml` file with your server details.
3. Run the deployment command:
   ```
   kamal deploy
   ```

## Docker

A Dockerfile is provided for containerization. To build and run the Docker image:

```
docker build -t coachbot .
docker run -e TELEGRAM_TOKEN=your_token -e CHAT_ID=your_chat_id ... coachbot
```
