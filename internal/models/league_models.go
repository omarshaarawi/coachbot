package models

import "time"

type LeagueMetadata struct {
	LeagueID             int
	Name                 string
	CurrentWeek          int
	CurrentScoringPeriod int
	SeasonID             int
	FirstWeek            int
	LastWeek             int
	IsActive             bool
	LastUpdated          time.Time
}

type TeamStanding struct {
	Rank          int
	TeamID        int
	TeamName      string
	Abbreviation  string
	Wins          int
	Losses        int
	Ties          int
	PointsFor     float64
	PointsAgainst float64
	WinPercentage float64
	PlayoffSeed   int
}

type CurrentScore struct {
	MatchID       int
	HomeTeamID    int
	AwayTeamID    int
	HomeScore     float64
	AwayScore     float64
	HomeProjected float64
	AwayProjected float64
	IsCompleted   bool
}

type WhoHasResult struct {
	PlayerName   string
	TeamName     string
	TeamID       int
	Found        bool
	PercentOwned float64
	Position     string
	ProTeam      string
	Points       float64
	IsProjected  bool
}

type PlayerToMonitor struct {
	Name         string
	Position     string
	InjuryStatus string
}

type TeamMonitorReport struct {
	TeamName string
	Players  []PlayerToMonitor
}

type PlayersToMonitorReport struct {
	Teams []TeamMonitorReport
}

type Matchup struct {
	HomeTeam  string
	AwayTeam  string
	HomeScore float64
	AwayScore float64
}

type Trophy struct {
	Category string
	Team     string
	Value    float64
}

type FinalScoreReport struct {
	Matchups []Matchup
	Trophies []Trophy
}

type CloseGame struct {
	HomeTeam  string
	AwayTeam  string
	HomeScore float64
	AwayScore float64
	Margin    float64
}
