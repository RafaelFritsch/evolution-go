package poll_service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/evolution-foundation/evolution-go/pkg/config"
	logger_wrapper "github.com/evolution-foundation/evolution-go/pkg/logger"
	"github.com/evolution-foundation/evolution-go/pkg/poll/model"
)

func TestSavePollVoteUsesConfiguredDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svc := &pollService{
		db:            db,
		loggerWrapper: newTestLoggerManager(t),
		queryTimeout:  time.Second,
	}

	mock.ExpectExec("INSERT INTO poll_votes").
		WithArgs(
			sqlmock.AnyArg(),
			"company",
			"instance",
			"poll",
			"chat",
			"vote",
			"user@s.whatsapp.net",
			"user",
			"User",
			"{A}",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = svc.SavePollVote(context.Background(), &model.PollVote{
		CompanyID:     "company",
		InstanceID:    "instance",
		PollMessageID: "poll",
		PollChatJid:   "chat",
		VoteMessageID: "vote",
		VoterJid:      "user@s.whatsapp.net",
		VoterName:     "User",
		SelectedOptions: []string{
			"A",
		},
		VotedAt:    time.Now(),
		ReceivedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("expected save to succeed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetPollResultsUsesConfiguredDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svc := &pollService{
		db:            db,
		loggerWrapper: newTestLoggerManager(t),
		queryTimeout:  time.Second,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "company_id", "instance_id", "poll_message_id", "poll_chat_jid",
		"vote_message_id", "voter_jid", "voter_phone", "voter_name",
		"selected_options", "voted_at", "received_at",
	}).AddRow(
		"id", "company", "instance", "poll", "chat",
		"vote", "user@s.whatsapp.net", "user", "User",
		"{A,B}", now, now,
	)

	mock.ExpectQuery("SELECT (.+) FROM poll_votes").
		WithArgs("poll", "instance").
		WillReturnRows(rows)

	results, err := svc.GetPollResults(context.Background(), "poll", "instance")
	if err != nil {
		t.Fatalf("expected get results to succeed: %v", err)
	}
	if results.TotalVotes != 1 {
		t.Fatalf("expected 1 vote, got %d", results.TotalVotes)
	}
	if results.OptionCounts["A"] != 1 || results.OptionCounts["B"] != 1 {
		t.Fatalf("unexpected option counts: %#v", results.OptionCounts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func newTestLoggerManager(t *testing.T) *logger_wrapper.LoggerManager {
	t.Helper()
	return logger_wrapper.NewLoggerManager(&config.Config{
		LogDirectory:  t.TempDir(),
		LogMaxSize:    1,
		LogMaxBackups: 1,
		LogMaxAge:     1,
		LogCompress:   false,
	})
}
