package server

import (
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"log"
	"math"
	"time"
)

type Retrier struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func NewRetrier(maxAttempts int, baseDelay time.Duration, maxDelay time.Duration) *Retrier {
	return &Retrier{MaxAttempts: maxAttempts, BaseDelay: baseDelay, MaxDelay: maxDelay}
}

func (r *Retrier) RunSQL(fn func() error) error {
	var err error
	for attempt := 0; attempt <= r.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetriableError(err) {
			return err
		}

		log.Printf("Attempt %d failed with retriable error: %v. Retrying...", attempt, err)

		delay := r.BaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
		if delay > r.MaxDelay {
			delay = r.MaxDelay
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("after %d attempts, last error: %w", r.MaxAttempts, err)
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionFailure:
			return true
		case pgerrcode.CannotConnectNow:
			return true
		case pgerrcode.UniqueViolation:
			return true
		default:
			return false
		}
	}

	return false
}
