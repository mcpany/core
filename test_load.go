package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// We won't use t.Parallel() since go-sqlmock cannot do MatchExpectationsInOrder(false).
// We will test each query sequentially by modifying `Load` to sequential for tests?
// Actually, `store.Load` runs concurrently. We must use `sqlmock`'s MatchAnyArgs or custom matching.
// OR we can just use `ExpectQuery(".*").WillReturnRows(...)` five times and let it match any.
// BUT `WillReturnRows` returns specific rows which must match the query!
// If we return users rows for upstream_services, it unmarshals successfully if fields match or ignore unknown.
// This is flaky.
