package espn

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/omarshaarawi/coachbot/internal/models"
)

type API struct {
	client *Client
}

func NewAPI(client *Client) *API {
	return &API{client: client}
}

func (a *API) GetLeagueMetadata() (*models.LeagueMetadata, error) {
	var espnResponse models.LeagueResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view": "mSettings",
	}

	if err := a.client.Get(endpoint, params, nil, &espnResponse); err != nil {
		return nil, fmt.Errorf("fetching league metadata: %w", err)
	}

	metadata := &models.LeagueMetadata{
		LeagueID:             espnResponse.ID,
		Name:                 espnResponse.Settings.Name,
		CurrentWeek:          espnResponse.Status.CurrentMatchupPeriod,
		CurrentScoringPeriod: espnResponse.ScoringPeriodID,
		SeasonID:             espnResponse.SeasonID,
		FirstWeek:            espnResponse.Status.FirstScoringPeriod,
		LastWeek:             espnResponse.Status.FinalScoringPeriod,
		IsActive:             espnResponse.Status.IsActive,
		LastUpdated:          time.Now(),
	}

	return metadata, nil
}

func (a *API) GetStandings() ([]models.TeamStanding, error) {
	var leagueResponse models.LeagueResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view": "mTeam",
	}

	if err := a.client.Get(endpoint, params, nil, &leagueResponse); err != nil {
		return nil, fmt.Errorf("fetching standings: %w", err)
	}

	standings := make([]models.TeamStanding, len(leagueResponse.Teams))
	for i, team := range leagueResponse.Teams {
		standings[i] = models.TeamStanding{
			TeamID:        team.ID,
			TeamName:      team.Name,
			Abbreviation:  team.Abbreviation,
			Wins:          team.Record.Overall.Wins,
			Losses:        team.Record.Overall.Losses,
			Ties:          team.Record.Overall.Ties,
			PointsFor:     team.Record.Overall.PointsFor,
			PointsAgainst: team.Record.Overall.PointsAgainst,
			WinPercentage: team.Record.Overall.Percentage,
			PlayoffSeed:   team.PlayoffSeed,
		}
	}

	sort.Slice(standings, func(i, j int) bool {
		if standings[i].WinPercentage != standings[j].WinPercentage {
			return standings[i].WinPercentage > standings[j].WinPercentage
		}
		return standings[i].PointsFor > standings[j].PointsFor
	})

	for i := range standings {
		standings[i].Rank = i + 1
	}

	return standings, nil
}

func (a *API) GetCurrentScores(week int) ([]models.Matchup, error) {
	var scoreboardResponse models.ScoreboardResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)

	params := map[string]string{
		"view": "mScoreboard",
	}

	filters := map[string]interface{}{
		"schedule": map[string]interface{}{
			"filterMatchupPeriodIds": map[string]interface{}{
				"value": []int{week},
			},
		},
	}

	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("error marshalling filters: %w", err)
	}

	headers := map[string]string{
		"x-fantasy-filter": string(filtersJSON),
	}

	if err := a.client.Get(endpoint, params, headers, &scoreboardResponse); err != nil {
		return nil, fmt.Errorf("fetching current scores: %w", err)
	}

	var matchups []models.Matchup

	for _, match := range scoreboardResponse.Schedule {
		homeScore, homeProjected := getScoreAndProjected(match.Home)
		awayScore, awayProjected := getScoreAndProjected(match.Away)

		matchup := models.Matchup{
			MatchID:       match.ID,
			HomeTeamID:    match.Home.TeamID,
			AwayTeamID:    match.Away.TeamID,
			HomeScore:     homeScore,
			AwayScore:     awayScore,
			HomeProjected: homeProjected,
			AwayProjected: awayProjected,
			IsCompleted:   match.Winner != "UNDECIDED",
		}

		matchups = append(matchups, matchup)
	}
	return matchups, nil
}

func getScoreAndProjected(teamScore models.TeamScore) (float64, float64) {
	score := teamScore.TotalPointsLive
	if score == 0 {
		score = teamScore.TotalPoints
	}
	projected := teamScore.TotalProjectedPointsLive
	return math.Round(score*100) / 100, math.Round(projected*100) / 100
}

func isCurrentMatch(match models.MatchupScore, currentPeriod int) bool {
	if len(match.Home.RosterForCurrentScoringPeriod.Entries) > 0 {
		playerStats := match.Home.RosterForCurrentScoringPeriod.Entries[0].PlayerPoolEntry.Player.Stats
		for _, stat := range playerStats {
			if stat.ScoringPeriodID == currentPeriod && stat.StatSourceID == 0 {
				return true
			}
		}
	}
	return false
}

func (a *API) WhoHas(playerName string, week int) (models.WhoHasResult, error) {
	var playerCardResp models.PlayerCardResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view":            "kona_playercard",
		"scoringPeriodId": fmt.Sprintf("%d", week),
	}

	filters := map[string]interface{}{
		"players": map[string]interface{}{
			"limit": 1000,
			"sortPercOwned": map[string]interface{}{
				"sortPriority": 1,
				"sortAsc":      false,
			},
			"sortDraftRanks": map[string]interface{}{
				"sortPriority": 100,
				"sortAsc":      true,
				"value":        "STANDARD",
			},
		},
	}

	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return models.WhoHasResult{}, fmt.Errorf("error marshalling filters: %w", err)
	}

	headers := map[string]string{
		"x-fantasy-filter": string(filtersJSON),
	}

	if err := a.client.Get(endpoint, params, headers, &playerCardResp); err != nil {
		return models.WhoHasResult{}, fmt.Errorf("fetching player info: %w", err)
	}

	lowerPlayerName := strings.ToLower(playerName)
	result := searchPlayers(playerCardResp.Players, lowerPlayerName, week)

	return result, nil
}

func searchPlayers(players []models.PlayerPoolEntry, playerName string, week int) models.WhoHasResult {
	var bestMatch *models.PlayerPoolEntry
	bestScore := -1
	threshold := 0.7

	for i, player := range players {
		fullName := strings.ToLower(player.Player.FullName)
		distance := fuzzy.LevenshteinDistance(strings.ToLower(playerName), fullName)
		maxLen := float64(max(len(playerName), len(fullName)))
		similarity := 1 - float64(distance)/maxLen

		if similarity > threshold && (bestScore == -1 || similarity > float64(bestScore)) {
			bestScore = int(similarity * 100)
			bestMatch = &players[i]
		}
	}

	if bestMatch != nil {
		teamName := getTeamName(bestMatch.OnTeamID)
		points, isProjected := getPlayerPoints(*bestMatch, week)

		return models.WhoHasResult{
			PlayerName:   bestMatch.Player.FullName,
			TeamName:     teamName,
			TeamID:       bestMatch.OnTeamID,
			Found:        true,
			PercentOwned: bestMatch.Player.Ownership.PercentOwned,
			Position:     getPositionString(bestMatch.Player.DefaultPositionID),
			ProTeam:      getProTeamString(bestMatch.Player.ProTeamID),
			Points:       points,
			IsProjected:  isProjected,
		}
	}

	return models.WhoHasResult{
		PlayerName: playerName,
		Found:      false,
	}
}

func getPlayerPoints(player models.PlayerPoolEntry, week int) (float64, bool) {
	currentScoringPeriod := week

	for _, stat := range player.Player.Stats {
		if stat.ScoringPeriodID == currentScoringPeriod {
			if stat.StatSourceID == 0 {
				return stat.AppliedTotal, false
			} else if stat.StatSourceID == 1 {
				return stat.AppliedTotal, true
			}
		}
	}

	return player.AppliedStatTotal, true
}

func getPositionString(positionID int) string {
	positions := map[int]string{
		1: "QB", 2: "RB", 3: "WR", 4: "TE", 5: "K", 16: "D/ST",
	}
	if pos, ok := positions[positionID]; ok {
		return pos
	}
	return "Unknown"
}

func getProTeamString(proTeamID int) string {
	teams := map[int]string{
		1: "ATL", 2: "BUF", 3: "CHI", 4: "CIN", 5: "CLE", 6: "DAL", 7: "DEN", 8: "DET",
		9: "GB", 10: "TEN", 11: "IND", 12: "KC", 13: "LV", 14: "LAR", 15: "MIA", 16: "MIN",
		17: "NE", 18: "NO", 19: "NYG", 20: "NYJ", 21: "PHI", 22: "ARI", 23: "PIT", 24: "LAC",
		25: "SF", 26: "SEA", 27: "TB", 28: "WSH", 29: "CAR", 30: "JAX", 33: "BAL", 34: "HOU",
	}

	if team, ok := teams[proTeamID]; ok {
		return team
	}

	return "Unknown"
}

func getTeamName(teamID int) string {
	teams := map[int]string{
		2: "Coach Dad",
		3: "Megatron's Ghurl",
		6: "Team InvincibleVince",
		5: "Beyond Cursed",
		1: "I Weigh Less Than Omar",
		4: "UGF Pandas",
	}

	name, ok := teams[teamID]
	if !ok {
		return "Unknown"
	}

	return name
}

func (a *API) GetPlayersToMonitor(week int) (models.PlayersToMonitorReport, error) {
	var leagueResponse models.LeagueResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view":            "mRoster",
		"scoringPeriodId": fmt.Sprintf("%d", week),
	}

	if err := a.client.Get(endpoint, params, nil, &leagueResponse); err != nil {
		return models.PlayersToMonitorReport{}, fmt.Errorf("fetching league rosters: %w", err)
	}

	report := models.PlayersToMonitorReport{}

	for _, team := range leagueResponse.Teams {
		teamReport := models.TeamMonitorReport{
			TeamName: getTeamName(team.ID),
		}

		for _, entry := range team.Roster.Entries {
			player := entry.PlayerPoolEntry.Player
			if isStartingLineup(entry.LineupSlotID) && isPlayerToMonitor(player.InjuryStatus) {
				teamReport.Players = append(teamReport.Players, models.PlayerToMonitor{
					Name:         player.FullName,
					Position:     getPositionString(player.DefaultPositionID),
					InjuryStatus: player.InjuryStatus,
				})
			}
		}

		if len(teamReport.Players) > 0 {
			report.Teams = append(report.Teams, teamReport)
		}
	}

	return report, nil
}

func isStartingLineup(slotID int) bool {
	startingSlots := map[int]bool{
		0:  true,  // QB
		2:  true,  // RB
		4:  true,  // WR
		6:  true,  // TE
		16: true,  // D/ST
		17: true,  // K
		20: false, // Bench
		21: false, // IR
		23: true,  // FLEX
	}
	return startingSlots[slotID]
}

func isPlayerToMonitor(status string) bool {
	return status == "QUESTIONABLE" || status == "DOUBTFUL" || status == "OUT"
}
