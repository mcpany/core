package resilience

import (
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRetry(t *testing.T) {
	t.Run("success_on_first_try", func(t *testing.T) {
		var attempts int
		work := func() error {
			attempts++
			return nil
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		retry := NewRetry(config)
		err := retry.Execute(work)
		require.NoError(t, err)
		require.Equal(t, 1, attempts)
	})

	t.Run("success_after_retries", func(t *testing.T) {
		var attempts int
		work := func() error {
			attempts++
			if attempts < 3 {
				return errors.New("transient error")
			}
			return nil
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(work)
		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("failure_after_all_retries", func(t *testing.T) {
		var attempts int
		work := func() error {
			attempts++
			return errors.New("persistent error")
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(work)
		require.Error(t, err)
		require.Equal(t, 4, attempts)
	})

	t.Run("permanent_error", func(t *testing.T) {
		var attempts int
		work := func() error {
			attempts++
			return &PermanentError{Err: errors.New("permanent error")}
		}

		retries := int32(3)
		config := &configv1.RetryConfig{}
		config.SetNumberOfRetries(retries)
		config.SetBaseBackoff(durationpb.New(1 * time.Millisecond))
		retry := NewRetry(config)
		err := retry.Execute(work)
		require.Error(t, err)
		require.Equal(t, 1, attempts)

		unwrappedErr := errors.Unwrap(err)
		require.NotNil(t, unwrappedErr)
		require.Equal(t, "permanent error", unwrappedErr.Error())
	})

	t.Run("default_backoff", func(t *testing.T) {
		config := &configv1.RetryConfig{}
		retry := NewRetry(config)
		require.Equal(t, time.Second, retry.config.GetBaseBackoff().AsDuration())
		require.Equal(t, 30*time.Second, retry.config.GetMaxBackoff().AsDuration())
	})

	t.Run("backoff_capping", func(t *testing.T) {
		config := &configv1.RetryConfig{}
		config.SetBaseBackoff(durationpb.New(2 * time.Second))
		config.SetMaxBackoff(durationpb.New(5 * time.Second))
		retry := NewRetry(config)
		require.Equal(t, 4*time.Second, retry.backoff(1))
		require.Equal(t, 5*time.Second, retry.backoff(2))
	})
}
