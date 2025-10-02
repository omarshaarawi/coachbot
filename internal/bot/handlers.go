package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/omarshaarawi/coachbot/internal/service"
)

type Handler struct {
	fantasyService *service.FantasyService
}

func NewHandler(fantasyService *service.FantasyService) *Handler {
	return &Handler{fantasyService: fantasyService}
}

func (h *Handler) HandleCommand(update tgbotapi.Update) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	command := strings.ToLower(update.Message.Command())
	args := update.Message.CommandArguments()
	msg.ParseMode = "Markdown"

	switch command {
	case "start":
		msg.Text = "Welcome to CoachBot! Use /help to see available commands."
	case "help":
		msg.Text = "Available commands:\n/scores - Get current scores\n/standings - Get league standings\n/team <team> - View team's roster and points\n/whohas <player> - Check which team has a player\n/monitor - Get players to monitor\n/finalscore - Get final score report\n/mondaynight - Get close games for Monday night\n/matchup - Get matchups for this week"
	case "scores":
		h.handleScores(&msg)
	case "standings":
		h.handleStandings(&msg)
	case "whohas":
		h.handleWhoHas(&msg, args)
	case "monitor":
		h.handlePlayersToMonitor(&msg)
	case "finalscore":
		h.handleFinalScore(&msg)
	case "mondaynight":
		h.handleMondayNightGames(&msg)
	case "matchup":
		h.handleMatchup(&msg)
	case "team":
		h.handleTeam(&msg, args)
	default:
		msg.Text = "Unknown command. Use /help to see available commands."
	}

	return msg
}

func (h *Handler) handleScores(msg *tgbotapi.MessageConfig) {
	scores, err := h.fantasyService.GetCurrentScores()
	if err != nil {
		msg.Text = fmt.Sprintf("Error fetching scores: %v", err)
	} else {
		msg.Text = scores
	}
}

func (h *Handler) handleStandings(msg *tgbotapi.MessageConfig) {
	standings, err := h.fantasyService.GetStandings()
	if err != nil {
		msg.Text = fmt.Sprintf("Error fetching standings: %v", err)
	} else {
		msg.Text = standings
	}
}

func (h *Handler) handleWhoHas(msg *tgbotapi.MessageConfig, args string) {
	if args == "" {
		msg.Text = "Please provide a player name. Usage: /whohas <player name>"
		return
	}
	result, err := h.fantasyService.WhoHas(args)
	if err != nil {
		msg.Text = fmt.Sprintf("Error checking who has player: %v", err)
	} else {
		msg.Text = result
	}
}

func (h *Handler) handlePlayersToMonitor(msg *tgbotapi.MessageConfig) {
	report, err := h.fantasyService.GetPlayersToMonitor()
	if err != nil {
		msg.Text = fmt.Sprintf("Error fetching players to monitor: %v", err)
	} else {
		msg.Text = report
	}
}

func (h *Handler) handleFinalScore(msg *tgbotapi.MessageConfig) {
	report, err := h.fantasyService.GetFinalScoreReport()
	if err != nil {
		msg.Text = fmt.Sprintf("Error generating final score report: %v", err)
	} else {
		msg.Text = report
	}
}

func (h *Handler) handleMondayNightGames(msg *tgbotapi.MessageConfig) {
	report, err := h.fantasyService.GetMondayNightCloseGames()
	if err != nil {
		msg.Text = fmt.Sprintf("Error generating Monday night close games report: %v", err)
	} else {
		msg.Text = report
	}
}

func (h *Handler) handleMatchup(msg *tgbotapi.MessageConfig) {
	report, err := h.fantasyService.GetMatchups()
	if err != nil {
		msg.Text = fmt.Sprintf("Error generating matchups report: %v", err)
	} else {
		msg.Text = report
	}
}

func (h *Handler) handleTeam(msg *tgbotapi.MessageConfig, args string) {
	if args == "" {
		msg.Text = "Please provide a team name. Usage: /team <team name>"
		return
	}
	result, err := h.fantasyService.GetTeamRoster(args)
	if err != nil {
		msg.Text = fmt.Sprintf("Error getting team roster: %v", err)
	} else {
		msg.Text = result
	}
}
