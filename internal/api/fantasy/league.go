package fantasy

import (
	"github.com/omarshaarawi/coachbot/internal/api/espn"
	"github.com/omarshaarawi/coachbot/internal/models"
)

type API struct {
	espnAPI *espn.API
}

func NewAPI(espnAPI *espn.API) *API {
	return &API{espnAPI: espnAPI}
}

func (a *API) GetLeagueMetadata() (*models.LeagueMetadata, error) {
	return a.espnAPI.GetLeagueMetadata()
}

func (a *API) GetStandings() ([]models.TeamStanding, error) {
	return a.espnAPI.GetStandings()
}

func (a *API) GetCurrentScores(week int) ([]models.Matchup, error) {
	return a.espnAPI.GetCurrentScores(week)
}

func (a *API) WhoHas(playerName string, week int) (models.WhoHasResult, error) {
	return a.espnAPI.WhoHas(playerName, week)
}

func (a *API) GetPlayersToMonitor(week int) (models.PlayersToMonitorReport, error) {
	return a.espnAPI.GetPlayersToMonitor(week)
}

func (a *API) GetTeamRoster(teamName string, week int) (models.TeamRoster, error) {
	return a.espnAPI.GetTeamRoster(teamName, week)
}
