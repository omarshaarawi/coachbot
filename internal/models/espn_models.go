package models

type LeagueResponse struct {
	ID              int      `json:"id"`
	ScoringPeriodID int      `json:"scoringPeriodId"`
	SeasonID        int      `json:"seasonId"`
	SegmentID       int      `json:"segmentId"`
	Status          Status   `json:"status"`
	Teams           []Team   `json:"teams"`
	Settings        Settings `json:"settings"`
}

type Settings struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type Status struct {
	CurrentMatchupPeriod int  `json:"currentMatchupPeriod"`
	FinalScoringPeriod   int  `json:"finalScoringPeriod"`
	FirstScoringPeriod   int  `json:"firstScoringPeriod"`
	IsActive             bool `json:"isActive"`
}

type Team struct {
	ID           int     `json:"id"`
	Abbreviation string  `json:"abbrev"`
	Name         string  `json:"name"`
	PlayoffSeed  int     `json:"playoffSeed"`
	Points       float64 `json:"points"`
	Roster       Roster  `json:"roster"`
	Record       Record  `json:"record"`
}

type Roster struct {
	Entries []RosterEntry `json:"entries"`
}

type Record struct {
	Overall RecordDetails `json:"overall"`
}

type RecordDetails struct {
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Ties          int     `json:"ties"`
	Percentage    float64 `json:"percentage"`
	PointsFor     float64 `json:"pointsFor"`
	PointsAgainst float64 `json:"pointsAgainst"`
}

type ScoreboardResponse struct {
	Schedule []MatchupScore `json:"schedule"`
}

type MatchupScore struct {
	ID     int       `json:"id"`
	Away   TeamScore `json:"away"`
	Home   TeamScore `json:"home"`
	Winner string    `json:"winner"`
}

type TeamScore struct {
	TeamID                        int             `json:"teamId"`
	TotalPoints                   float64         `json:"totalPoints"`
	TotalPointsLive               float64         `json:"totalPointsLive"`
	TotalProjectedPointsLive      float64         `json:"totalProjectedPointsLive"`
	RosterForCurrentScoringPeriod RosterForPeriod `json:"rosterForCurrentScoringPeriod"`
}

type RosterForPeriod struct {
	Entries []RosterEntry `json:"entries"`
}

type RosterEntry struct {
	PlayerPoolEntry PlayerPoolEntry `json:"playerPoolEntry"`
	LineupSlotID    int             `json:"lineupSlotId"`
}

type PlayerCardResponse struct {
	Players []PlayerPoolEntry `json:"players"`
}

type PlayerPoolEntry struct {
	ID               int     `json:"id"`
	OnTeamID         int     `json:"onTeamId"`
	Player           Player  `json:"player"`
	AppliedStatTotal float64 `json:"appliedStatTotal"`
}

type Player struct {
	ID                int       `json:"id"`
	FullName          string    `json:"fullName"`
	DefaultPositionID int       `json:"defaultPositionId"`
	ProTeamID         int       `json:"proTeamId"`
	Ownership         Ownership `json:"ownership"`
	Stats             []Stat    `json:"stats"`
	InjuryStatus      string    `json:"injuryStatus"`
}

type Ownership struct {
	PercentOwned float64 `json:"percentOwned"`
}

type Stat struct {
	StatSourceID    int                `json:"statSourceId"`
	ScoringPeriodID int                `json:"scoringPeriodId"`
	AppliedTotal    float64            `json:"appliedTotal"`
	AppliedStats    map[string]float64 `json:"appliedStats"`
}
