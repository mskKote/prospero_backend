package lib

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
)

var logger = logging.GetLogger()

func HandlePgErr(err error) error {
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			logger.Error(pgErr.Message, zap.Error(pgErr))
			return pgErr
		}
		return err
	}
	return nil
}
