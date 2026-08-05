package whatsmeow_service

import (
	"context"
	"errors"
	"testing"

	"go.mau.fi/whatsmeow/appstate"
)

type fakeNCTSaltStore struct {
	salt []byte
	err  error
	gets int
}

func (s *fakeNCTSaltStore) GetNCTSalt(context.Context) ([]byte, error) {
	s.gets++
	if s.err != nil {
		return nil, s.err
	}
	return s.salt, nil
}

type fakeAppStateFetcher struct {
	store           *fakeNCTSaltStore
	err             error
	calls           int
	name            appstate.WAPatchName
	fullSync        bool
	onlyIfNotSynced bool
}

func (f *fakeAppStateFetcher) FetchAppState(_ context.Context, name appstate.WAPatchName, fullSync, onlyIfNotSynced bool) error {
	f.calls++
	f.name = name
	f.fullSync = fullSync
	f.onlyIfNotSynced = onlyIfNotSynced
	if f.err != nil {
		return f.err
	}
	if f.store != nil {
		f.store.salt = []byte("synced-nct-salt")
	}
	return nil
}

type fakeNCTSaltLogger struct{}

func (fakeNCTSaltLogger) LogInfo(string, ...interface{})  {}
func (fakeNCTSaltLogger) LogWarn(string, ...interface{})  {}
func (fakeNCTSaltLogger) LogError(string, ...interface{}) {}

func TestNCTSaltBootstrapFetchesRegularHighWhenSaltIsMissing(t *testing.T) {
	store := &fakeNCTSaltStore{}
	fetcher := &fakeAppStateFetcher{store: store}
	bootstrapper := &nctSaltBootstrapper{}

	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})

	if fetcher.calls != 1 {
		t.Fatalf("expected one FetchAppState call, got %d", fetcher.calls)
	}
	if fetcher.name != appstate.WAPatchRegularHigh {
		t.Fatalf("expected regular_high fetch, got %s", fetcher.name)
	}
	if !fetcher.fullSync {
		t.Fatal("expected fullSync=true")
	}
	if fetcher.onlyIfNotSynced {
		t.Fatal("expected onlyIfNotSynced=false")
	}
	if store.gets != 2 {
		t.Fatalf("expected salt to be read before and after fetch, got %d reads", store.gets)
	}
}

func TestNCTSaltBootstrapRunsAtMostOnceWhenSaltStaysMissing(t *testing.T) {
	store := &fakeNCTSaltStore{}
	fetcher := &fakeAppStateFetcher{}
	bootstrapper := &nctSaltBootstrapper{}

	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})
	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})

	if fetcher.calls != 1 {
		t.Fatalf("expected one FetchAppState call, got %d", fetcher.calls)
	}
}

func TestNCTSaltBootstrapSkipsFetchWhenSaltAlreadyExists(t *testing.T) {
	store := &fakeNCTSaltStore{salt: []byte("existing-nct-salt")}
	fetcher := &fakeAppStateFetcher{}
	bootstrapper := &nctSaltBootstrapper{}

	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})

	if fetcher.calls != 0 {
		t.Fatalf("expected no FetchAppState calls, got %d", fetcher.calls)
	}
	if store.gets != 1 {
		t.Fatalf("expected one salt read, got %d", store.gets)
	}
}

func TestNCTSaltBootstrapHandlesFetchErrorWithoutRetryLoop(t *testing.T) {
	store := &fakeNCTSaltStore{}
	fetcher := &fakeAppStateFetcher{err: errors.New("fetch failed")}
	bootstrapper := &nctSaltBootstrapper{}

	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})
	bootstrapper.Ensure(context.Background(), "instance-1", store, fetcher, fakeNCTSaltLogger{})

	if fetcher.calls != 1 {
		t.Fatalf("expected one FetchAppState call after error, got %d", fetcher.calls)
	}
}
