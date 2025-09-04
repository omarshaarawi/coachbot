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

// func isCurrentMatch(match models.MatchupScore, currentPeriod int) bool {
// 	if len(match.Home.RosterForCurrentScoringPeriod.Entries) > 0 {
// 		playerStats := match.Home.RosterForCurrentScoringPeriod.Entries[0].PlayerPoolEntry.Player.Stats
// 		for _, stat := range playerStats {
// 			if stat.ScoringPeriodID == currentPeriod && stat.StatSourceID == 0 {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

func (a *API) WhoHas(playerName string, week int) (models.WhoHasResult, error) {
	var leagueResponse models.LeagueResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view":            "mRoster",
		"scoringPeriodId": fmt.Sprintf("%d", week),
	}

	if err := a.client.Get(endpoint, params, nil, &leagueResponse); err != nil {
		return models.WhoHasResult{}, fmt.Errorf("fetching league rosters: %w", err)
	}

	var allPlayers []models.PlayerPoolEntry
	for _, team := range leagueResponse.Teams {
		for _, entry := range team.Roster.Entries {
			allPlayers = append(allPlayers, entry.PlayerPoolEntry)
		}
	}

	return searchPlayers(leagueResponse.Teams, allPlayers, playerName, week), nil
}

func searchPlayers(teams []models.Team, players []models.PlayerPoolEntry, playerName string, week int) models.WhoHasResult {
	var bestMatch *models.PlayerPoolEntry
	var bestMatchEntry *models.RosterEntry
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

			// Find the roster entry for this player
			for _, team := range teams {
				for _, entry := range team.Roster.Entries {
					if entry.PlayerPoolEntry.ID == bestMatch.ID {
						bestMatchEntry = &entry
						break
					}
				}
				if bestMatchEntry != nil {
					break
				}
			}
		}
	}

	if bestMatch != nil {
		teamName := getTeamName(bestMatch.OnTeamID)
		points, isProjected := getPlayerPoints(*bestMatch, week)

		lineupSlot := "Unknown"
		if bestMatchEntry != nil {
			lineupSlot = getLineupSlotString(bestMatchEntry.LineupSlotID)
		}

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
			LineupSlot:   lineupSlot,
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
		3: "Stairway to Evans",
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

func (a *API) GetTeamRoster(teamName string, week int) (models.TeamRoster, error) {
	var leagueResponse models.LeagueResponse
	var scoreboardResponse models.ScoreboardResponse
	endpoint := fmt.Sprintf("/seasons/%s/segments/0/leagues/%s", a.client.Config.Year, a.client.Config.LeagueID)
	params := map[string]string{
		"view":            "mRoster",
		"scoringPeriodId": fmt.Sprintf("%d", week),
	}

	if err := a.client.Get(endpoint, params, nil, &leagueResponse); err != nil {
		return models.TeamRoster{}, fmt.Errorf("fetching league rosters: %w", err)
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
		return models.TeamRoster{}, fmt.Errorf("error marshalling filters: %w", err)
	}

	headers := map[string]string{
		"x-fantasy-filter": string(filtersJSON),
	}

	if err := a.client.Get(endpoint, map[string]string{"view": "mScoreboard"}, headers, &scoreboardResponse); err != nil {
		return models.TeamRoster{}, fmt.Errorf("fetching scoreboard: %w", err)
	}

	var bestMatch *models.Team
	bestScore := -1
	threshold := 0.6

	for i, team := range leagueResponse.Teams {
		currentTeamName := getTeamName(team.ID)
		distance := fuzzy.LevenshteinDistance(strings.ToLower(teamName), strings.ToLower(currentTeamName))
		maxLen := float64(max(len(teamName), len(currentTeamName)))
		similarity := 1 - float64(distance)/maxLen

		if similarity > threshold && (bestScore == -1 || similarity > float64(bestScore)) {
			bestScore = int(similarity * 100)
			bestMatch = &leagueResponse.Teams[i]
		}
	}

	if bestMatch == nil {
		return models.TeamRoster{}, fmt.Errorf("team not found: %s", teamName)
	}

	roster := models.TeamRoster{
		TeamName: getTeamName(bestMatch.ID),
		Players:  make([]models.RosterPlayer, 0),
	}

	var starters []models.RosterPlayer
	var bench []models.RosterPlayer

	byeWeeks, err := a.GetProSchedule()
	if err != nil {
		return models.TeamRoster{}, fmt.Errorf("fetching pro schedule: %w", err)
	}

	for _, entry := range bestMatch.Roster.Entries {
		player := entry.PlayerPoolEntry.Player
		points, _ := getPlayerPoints(entry.PlayerPoolEntry, week)

		pointsDisplay := "TBD"
		if entry.LineupSlotID == 21 || player.InjuryStatus == "INJURY_RESERVE" {
			pointsDisplay = "IR"
		} else if byeWeek, ok := byeWeeks[player.ProTeamID]; ok && byeWeek == week {
			pointsDisplay = "BYE"
		} else {
			hasActualStats := false

			for _, stat := range player.Stats {
				if stat.ScoringPeriodID == week {
					if stat.StatSourceID == 0 {
						hasActualStats = true
						pointsDisplay = fmt.Sprintf("%.2f", stat.AppliedTotal)
						break
					}
				}
			}

			if !hasActualStats {
				pointsDisplay = "TBD"
			}
		}

		rosterPlayer := models.RosterPlayer{
			Name:         player.FullName,
			Position:     getPositionString(player.DefaultPositionID),
			Points:       points,
			PointsLabel:  pointsDisplay,
			IsStarter:    isStartingLineup(entry.LineupSlotID),
			LineupSlot:   getLineupSlotString(entry.LineupSlotID),
			InjuryStatus: player.InjuryStatus,
		}

		if isStartingLineup(entry.LineupSlotID) {
			starters = append(starters, rosterPlayer)
		} else {
			bench = append(bench, rosterPlayer)
		}
	}

	sort.Slice(starters, func(i, j int) bool {
		order := map[string]int{
			"QB":   1,
			"RB":   2,
			"WR":   3,
			"TE":   4,
			"FLEX": 5,
			"D/ST": 6,
			"K":    7,
		}
		return order[starters[i].Position] < order[starters[j].Position]
	})

	roster.Players = append(roster.Players, starters...)
	roster.Players = append(roster.Players, bench...)

	return roster, nil
}

func getLineupSlotString(slotID int) string {
	switch slotID {
	case 0:
		return "QB"
	case 2:
		return "RB"
	case 4:
		return "WR"
	case 6:
		return "TE"
	case 16:
		return "D/ST"
	case 17:
		return "K"
	case 20:
		return "Bench"
	case 21:
		return "IR"
	case 23:
		return "FLEX"
	default:
		return "Unknown"
	}
}

type ProTeamInfo struct {
	ID      int    `json:"id"`
	Abbrev  string `json:"abbrev"`
	ByeWeek int    `json:"byeWeek"`
	Name    string `json:"name"`
}

func (a *API) GetProSchedule() (map[int]int, error) {
	var scheduleResponse struct {
		Settings struct {
			ProTeams []ProTeamInfo `json:"proTeams"`
		} `json:"settings"`
	}

	endpoint := fmt.Sprintf("/seasons/%s", a.client.Config.Year)
	params := map[string]string{
		"view": "proTeamSchedules_wl",
	}

	if err := a.client.Get(endpoint, params, nil, &scheduleResponse); err != nil {
		return nil, fmt.Errorf("fetching pro schedule: %w", err)
	}

	byeWeeks := make(map[int]int)
	for _, team := range scheduleResponse.Settings.ProTeams {
		if team.ByeWeek > 0 {
			byeWeeks[team.ID] = team.ByeWeek
		}
	}

	return byeWeeks, nil
}
