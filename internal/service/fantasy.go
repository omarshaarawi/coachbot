package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/omarshaarawi/coachbot/internal/api/fantasy"
	"github.com/omarshaarawi/coachbot/internal/models"
	"github.com/omarshaarawi/coachbot/internal/repository/memory"
)

type FantasyService struct {
	api  *fantasy.API
	repo *memory.Repository
}

func NewFantasyService(api *fantasy.API, repo *memory.Repository) *FantasyService {
	return &FantasyService{api: api, repo: repo}
}

func (s *FantasyService) GetCurrentWeek() (int, error) {
	metadata, err := s.getLeagueMetadata()
	if err != nil {
		return 0, err
	}
	return metadata.CurrentWeek, nil
}

func (s *FantasyService) getLeagueMetadata() (*models.LeagueMetadata, error) {
	metadata := s.repo.GetMetadata()
	if metadata == nil || time.Since(metadata.LastUpdated) > 24*time.Hour {
		newMetadata, err := s.api.GetLeagueMetadata()
		if err != nil {
			return nil, err
		}
		s.repo.SaveMetadata(newMetadata)
		return newMetadata, nil
	}
	return metadata, nil
}

func (s *FantasyService) GetStandings() (string, error) {
	standings, err := s.api.GetStandings()
	if err != nil {
		return "", fmt.Errorf("error fetching standings: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("🏆 *Current Standings*\n\n")
	for _, team := range standings {
		sb.WriteString(fmt.Sprintf("%d. *%s* - %d-%d-%d, %.2f PF, %.2f PA\n",
			team.Rank,
			team.TeamName,
			team.Wins,
			team.Losses,
			team.Ties,
			team.PointsFor,
			team.PointsAgainst))
	}

	return sb.String(), nil
}

func (s *FantasyService) GetCurrentScores() (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	scores, err := s.api.GetCurrentScores(week)
	if err != nil {
		return "", fmt.Errorf("error fetching current scores: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("🏈 Current Scores (Live / Projected):\n\n")

	for _, score := range scores {
		homeTeam := getTeamName(score.HomeTeamID)
		awayTeam := getTeamName(score.AwayTeamID)

		sb.WriteString(fmt.Sprintf("*%s* %.2f (%.2f) - %.2f (%.2f) *%s*\n",
			homeTeam, score.HomeScore, score.HomeProjected,
			score.AwayScore, score.AwayProjected, awayTeam))

		if score.IsCompleted {
			sb.WriteString("(Final)\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

var teamIDToName = map[int]string{
	2: "Coach Dad",
	3: "Megatron's Ghurl",
	6: "Team InvincibleVince",
	5: "Beyond Cursed",
	1: "I Weigh Less Than Omar",
	4: "UGF Pandas",
}

func getTeamName(teamID int) string {
	name, ok := teamIDToName[teamID]
	if !ok {
		return "Unknown"
	}
	return name
}

func (s *FantasyService) WhoHas(playerName string) (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	result, err := s.api.WhoHas(playerName, week)
	if err != nil {
		return "", fmt.Errorf("error checking who has player: %w", err)
	}

	if !result.Found {
		return fmt.Sprintf("🔍 No player found matching '%s'.", playerName), nil
	}

	var status string
	if result.TeamID != 0 {
		status = fmt.Sprintf("on the roster of *%s*", result.TeamName)
	} else {
		status = "a free agent"
	}

	pointsType := "Points"
	if result.IsProjected {
		pointsType = "Projected Points"
	}

	return fmt.Sprintf("🔍 *Player Information*\n\n%s (%s, %s) is %s\nOwned in %.1f%% of leagues\n%s: %.2f",
		result.PlayerName, result.Position, result.ProTeam, status, result.PercentOwned, pointsType, result.Points), nil
}

func (s *FantasyService) GetPlayersToMonitor() (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	report, err := s.api.GetPlayersToMonitor(week)
	if err != nil {
		return "", fmt.Errorf("error fetching players to monitor: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("🚑 *Players to Monitor*\n\n")

	if len(report.Teams) == 0 {
		sb.WriteString("No players to monitor at this time.")
		return sb.String(), nil
	}

	for _, team := range report.Teams {
		sb.WriteString(fmt.Sprintf("*%s:*\n", team.TeamName))
		for _, player := range team.Players {
			sb.WriteString(fmt.Sprintf("- %s %s - %s\n", player.Position, player.Name, player.InjuryStatus))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (s *FantasyService) GetFinalScoreReport() (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	currentScores, err := s.api.GetCurrentScores(week)
	if err != nil {
		return "", fmt.Errorf("error fetching matchups: %w", err)
	}

	report := processScores(currentScores)
	return formatFinalScoreReport(report), nil
}

func processScores(scores []models.CurrentScore) models.FinalScoreReport {
	var report models.FinalScoreReport
	report.Matchups = make([]models.Matchup, len(scores))

	var highScore, lowScore float64
	var biggestWin, closestWin float64
	var highScoreTeam, lowScoreTeam, biggestWinTeam, closestWinTeam string

	highScore = -math.MaxFloat64
	lowScore = math.MaxFloat64
	biggestWin = -math.MaxFloat64
	closestWin = math.MaxFloat64

	for i, score := range scores {
		homeTeam := getTeamName(score.HomeTeamID)
		awayTeam := getTeamName(score.AwayTeamID)

		report.Matchups[i] = models.Matchup{
			HomeTeam:  homeTeam,
			AwayTeam:  awayTeam,
			HomeScore: score.HomeScore,
			AwayScore: score.AwayScore,
		}

		// High Score
		if score.HomeScore > highScore {
			highScore = score.HomeScore
			highScoreTeam = homeTeam
		}
		if score.AwayScore > highScore {
			highScore = score.AwayScore
			highScoreTeam = awayTeam
		}

		// Low Score
		if score.HomeScore < lowScore {
			lowScore = score.HomeScore
			lowScoreTeam = homeTeam
		}
		if score.AwayScore < lowScore {
			lowScore = score.AwayScore
			lowScoreTeam = awayTeam
		}

		// Biggest Win and Closest Win
		scoreDiff := math.Abs(score.HomeScore - score.AwayScore)
		if scoreDiff > biggestWin {
			biggestWin = scoreDiff
			if score.HomeScore > score.AwayScore {
				biggestWinTeam = homeTeam
			} else {
				biggestWinTeam = awayTeam
			}
		}
		if scoreDiff < closestWin {
			closestWin = scoreDiff
			if score.HomeScore > score.AwayScore {
				closestWinTeam = homeTeam
			} else {
				closestWinTeam = awayTeam
			}
		}
	}

	report.Trophies = []models.Trophy{
		{Category: "High Score", Team: highScoreTeam, Value: highScore},
		{Category: "Low Score", Team: lowScoreTeam, Value: lowScore},
		{Category: "Biggest Win", Team: biggestWinTeam, Value: biggestWin},
		{Category: "Closest Win", Team: closestWinTeam, Value: closestWin},
	}

	return report
}

func formatFinalScoreReport(report models.FinalScoreReport) string {
	var sb strings.Builder

	sb.WriteString("📊 *Final Scores:*\n\n")

	sort.Slice(report.Matchups, func(i, j int) bool {
		totalScoreI := report.Matchups[i].HomeScore + report.Matchups[i].AwayScore
		totalScoreJ := report.Matchups[j].HomeScore + report.Matchups[j].AwayScore
		return totalScoreI > totalScoreJ
	})

	for _, m := range report.Matchups {
		sb.WriteString(fmt.Sprintf("%s %.2f - %.2f %s\n", m.HomeTeam, m.HomeScore, m.AwayScore, m.AwayTeam))
	}

	sb.WriteString("\n🏆 *Trophies:*\n")
	for _, t := range report.Trophies {
		switch t.Category {
		case "High Score":
			sb.WriteString(fmt.Sprintf("Highest Score: %s (%.2f)\n", t.Team, t.Value))
		case "Low Score":
			sb.WriteString(fmt.Sprintf("Lowest Score: %s (%.2f)\n", t.Team, t.Value))
		case "Biggest Win":
			sb.WriteString(fmt.Sprintf("Biggest Win: %s (Margin: %.2f)\n", t.Team, t.Value))
		case "Closest Win":
			sb.WriteString(fmt.Sprintf("Closest Win: %s (Margin: %.2f)\n", t.Team, t.Value))
		}
	}

	return sb.String()
}

func (s *FantasyService) GetMondayNightCloseGames() (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	currentScores, err := s.api.GetCurrentScores(week)
	if err != nil {
		return "", fmt.Errorf("error fetching current scores: %w", err)
	}

	closeGames := findCloseGames(currentScores)
	return formatMondayNightCloseGames(closeGames), nil
}

func findCloseGames(scores []models.CurrentScore) []models.CloseGame {
	var closeGames []models.CloseGame

	for _, score := range scores {
		margin := math.Abs(score.HomeScore - score.AwayScore)
		if margin <= 16 {
			closeGames = append(closeGames, models.CloseGame{
				HomeTeam:  getTeamName(score.HomeTeamID),
				AwayTeam:  getTeamName(score.AwayTeamID),
				HomeScore: score.HomeScore,
				AwayScore: score.AwayScore,
				Margin:    margin,
			})
		}
	}

	sort.Slice(closeGames, func(i, j int) bool {
		return closeGames[i].Margin < closeGames[j].Margin
	})

	return closeGames
}

func formatMondayNightCloseGames(closeGames []models.CloseGame) string {
	var sb strings.Builder

	sb.WriteString("🏈 *Monday Night Watch List*\n\n")

	if len(closeGames) == 0 {
		sb.WriteString("No close games this week. All outcomes are likely decided.")
		return sb.String()
	}

	for _, game := range closeGames {
		sb.WriteString(fmt.Sprintf("%s %.2f - %.2f %s (Margin: %.2f)\n",
			game.HomeTeam, game.HomeScore, game.AwayScore, game.AwayTeam, game.Margin))
	}

	return sb.String()
}

func (s *FantasyService) GetMatchups() (string, error) {
	week, err := s.GetCurrentWeek()
	if err != nil {
		return "", fmt.Errorf("error fetching current week: %w", err)
	}

	currentScores, err := s.api.GetCurrentScores(week)
	if err != nil {
		return "", fmt.Errorf("error fetching current scores: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("🏈 *Matchups*\n\n")

	for _, score := range currentScores {
		homeTeam := getTeamName(score.HomeTeamID)
		awayTeam := getTeamName(score.AwayTeamID)

		sb.WriteString(fmt.Sprintf("*%s* %.2f - %.2f *%s*\n",
			homeTeam, score.HomeScore, score.AwayScore, awayTeam))
	}

	return sb.String(), nil
}