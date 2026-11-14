package pgquery

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

func TestQueryError(t *testing.T) {
	t.Run("database error becomes a QueryError", func(t *testing.T) {
		err := queryError(context.Background(), fmt.Errorf("wrapped: %w", &pgconn.PgError{
			Code:     "42601",
			Message:  `syntax error at or near "SELEC"`,
			Hint:     "check the spelling",
			Position: 1,
		}))

		var queryErr *domain.QueryError
		require.ErrorAs(t, err, &queryErr)
		assert.Equal(t, &domain.QueryError{
			SQLState: "42601",
			Message:  `syntax error at or near "SELEC"`,
			Hint:     "check the spelling",
			Position: 1,
		}, queryErr)
	})

	t.Run("expired statement deadline is a timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		<-ctx.Done()

		// pgx surfaces the cancellation as a 57014 error from PostgreSQL.
		err := queryError(ctx, &pgconn.PgError{Code: "57014", Message: "canceling statement due to user request"})

		assert.ErrorIs(t, err, domain.ErrQueryTimeout)
	})

	t.Run("other errors pass through", func(t *testing.T) {
		original := errors.New("connection reset")

		assert.Same(t, original, queryError(context.Background(), original))
	})
}
