package scheduler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/omarshaarawi/coachbot/internal/service"
)

type Scheduler struct {
	s              gocron.Scheduler
	fantasyService *service.FantasyService
	sendMessage    func(string) error
}

func NewScheduler(fantasyService *service.FantasyService, sendMessage func(string) error) (*Scheduler, error) {
	location, err := time.LoadLocation("America/Chicago") // CDT
	if err != nil {
		slog.Error("Failed to load location", "error", err)
	}

	s, err := gocron.NewScheduler(
		gocron.WithLocation(location),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		s:              s,
		fantasyService: fantasyService,
		sendMessage:    sendMessage,
	}, nil
}

func (s *Scheduler) Start() error {
	var err error

	// Close Scores - Monday 18:30 EDT (17:30 CDT)
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday), gocron.NewAtTimes(gocron.NewAtTime(17, 30, 0))),
		gocron.NewTask(s.sendCloseScores),
	)
	if err != nil {
		return fmt.Errorf("failed to create close scores job: %w", err)
	}

	// Scoreboard - Monday, Tuesday, Friday 7:30 CDT
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Monday, time.Tuesday, time.Friday), gocron.NewAtTimes(gocron.NewAtTime(7, 30, 0))),
		gocron.NewTask(s.sendScoreboard),
	)
	if err != nil {
		return fmt.Errorf("failed to create scoreboard job: %w", err)
	}

	// Trophies - Tuesday 7:30 CDT
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Tuesday), gocron.NewAtTimes(gocron.NewAtTime(7, 30, 0))),
		gocron.NewTask(s.sendTrophies),
	)
	if err != nil {
		return fmt.Errorf("failed to create trophies job: %w", err)
	}

	// Current standings - Wednesday 7:30 CDT
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Wednesday), gocron.NewAtTimes(gocron.NewAtTime(7, 30, 0))),
		gocron.NewTask(s.sendStandings),
	)
	if err != nil {
		return fmt.Errorf("failed to create standings job: %w", err)
	}

	// Matchups - Thursday 19:30 EDT (18:30 CDT)
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Thursday), gocron.NewAtTimes(gocron.NewAtTime(18, 30, 0))),
		gocron.NewTask(s.sendMatchups),
	)
	if err != nil {
		return fmt.Errorf("failed to create matchups job: %w", err)
	}

	// Players to Monitor report - Sunday 7:30 CDT
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Sunday), gocron.NewAtTimes(gocron.NewAtTime(7, 30, 0))),
		gocron.NewTask(s.sendPlayersToMonitor),
	)
	if err != nil {
		return fmt.Errorf("failed to create players to monitor job: %w", err)
	}

	// Scoreboard - Sunday 16:00 and 20:00 EDT (15:00 and 19:00 CDT)
	_, err = s.s.NewJob(
		gocron.WeeklyJob(1, gocron.NewWeekdays(time.Sunday), gocron.NewAtTimes(gocron.NewAtTime(15, 0, 0), gocron.NewAtTime(19, 0, 0))),
		gocron.NewTask(s.sendScoreboard),
	)
	if err != nil {
		return fmt.Errorf("failed to create Sunday scoreboard job: %w", err)
	}

	s.s.Start()
	return nil
}

func (s *Scheduler) Stop() error {
	return s.s.Shutdown()
}

func (s *Scheduler) sendCloseScores() {
	report, err := s.fantasyService.GetMondayNightCloseGames()
	if err != nil {
		slog.Error("Failed to get close games", "error", err)
		return
	}
	s.sendMessage(report)
}

func (s *Scheduler) sendScoreboard() {
	scores, err := s.fantasyService.GetCurrentScores()
	if err != nil {
		slog.Error("Failed to get current scores", "error", err)
		return
	}
	s.sendMessage(scores)
}

func (s *Scheduler) sendTrophies() {
	report, err := s.fantasyService.GetFinalScoreReport()
	if err != nil {
		slog.Error("Failed to get final score report", "error", err)
		return
	}
	s.sendMessage(report)
}

func (s *Scheduler) sendStandings() {
	standings, err := s.fantasyService.GetStandings()
	if err != nil {
		slog.Error("Failed to get standings", "error", err)
		return
	}
	s.sendMessage(standings)
}

func (s *Scheduler) sendMatchups() {
	matchups, err := s.fantasyService.GetMatchups()
	if err != nil {
		slog.Error("Failed to get matchups", "error", err)
		return
	}
	s.sendMessage(matchups)
}

func (s *Scheduler) sendPlayersToMonitor() {
	report, err := s.fantasyService.GetPlayersToMonitor()
	if err != nil {
		slog.Error("Failed to get players to monitor", "error", err)
		return
	}
	s.sendMessage(report)
}
