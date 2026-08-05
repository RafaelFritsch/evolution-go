package whatsmeow_service

import (
	"context"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/appstate"
)

const nctSaltBootstrapTimeout = 30 * time.Second

type nctSaltStore interface {
	GetNCTSalt(ctx context.Context) ([]byte, error)
}

type nctSaltAppStateFetcher interface {
	FetchAppState(ctx context.Context, name appstate.WAPatchName, fullSync, onlyIfNotSynced bool) error
}

type nctSaltLogger interface {
	LogInfo(format string, args ...interface{})
	LogWarn(format string, args ...interface{})
	LogError(format string, args ...interface{})
}

type nctSaltBootstrapper struct {
	mu        sync.Mutex
	attempted bool
}

func (b *nctSaltBootstrapper) Ensure(
	ctx context.Context,
	instanceID string,
	saltStore nctSaltStore,
	fetcher nctSaltAppStateFetcher,
	logger nctSaltLogger,
) {
	if saltStore == nil || fetcher == nil {
		logNCTSaltWarn(logger, "[%s] NCT salt bootstrap skipped: store or app-state fetcher is unavailable", instanceID)
		return
	}

	salt, err := saltStore.GetNCTSalt(ctx)
	if err != nil {
		logNCTSaltWarn(logger, "[%s] NCT salt bootstrap skipped: failed to read stored salt: %v", instanceID, err)
		return
	}
	if len(salt) > 0 {
		logNCTSaltInfo(logger, "[%s] NCT salt already stored, skipping bootstrap", instanceID)
		return
	}

	if !b.markAttempted() {
		logNCTSaltInfo(logger, "[%s] NCT salt bootstrap already attempted in this process, skipping", instanceID)
		return
	}

	logNCTSaltWarn(logger, "[%s] NCT salt missing, forcing regular_high app-state sync", instanceID)
	if err = fetcher.FetchAppState(ctx, appstate.WAPatchRegularHigh, true, false); err != nil {
		logNCTSaltError(logger, "[%s] NCT salt bootstrap failed while fetching regular_high: %v", instanceID, err)
		return
	}

	salt, err = saltStore.GetNCTSalt(ctx)
	if err != nil {
		logNCTSaltWarn(logger, "[%s] NCT salt bootstrap completed but salt could not be re-read: %v", instanceID, err)
		return
	}
	if len(salt) == 0 {
		logNCTSaltWarn(logger, "[%s] NCT salt still missing after regular_high sync; this account may not have server-side salt provisioned", instanceID)
		return
	}
	logNCTSaltInfo(logger, "[%s] NCT salt stored after regular_high sync", instanceID)
}

func (b *nctSaltBootstrapper) markAttempted() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.attempted {
		return false
	}
	b.attempted = true
	return true
}

func (mycli *MyClient) ensureNCTSaltSyncedAsync() {
	if mycli == nil || mycli.WAClient == nil || mycli.WAClient.Store == nil || mycli.WAClient.Store.ID == nil {
		return
	}

	var logger nctSaltLogger
	if mycli.loggerWrapper != nil {
		logger = mycli.loggerWrapper.GetLogger(mycli.userID)
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), nctSaltBootstrapTimeout)
		defer cancel()

		mycli.nctSaltBootstrap.Ensure(ctx, mycli.userID, mycli.WAClient.Store.NCTSalt, mycli.WAClient, logger)
	}()
}

func logNCTSaltInfo(logger nctSaltLogger, format string, args ...interface{}) {
	if logger != nil {
		logger.LogInfo(format, args...)
	}
}

func logNCTSaltWarn(logger nctSaltLogger, format string, args ...interface{}) {
	if logger != nil {
		logger.LogWarn(format, args...)
	}
}

func logNCTSaltError(logger nctSaltLogger, format string, args ...interface{}) {
	if logger != nil {
		logger.LogError(format, args...)
	}
}
