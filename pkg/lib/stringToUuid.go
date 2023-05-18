package lib

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func StringToUUID(s string) pgtype.UUID {
	uuid := pgtype.UUID{}
	err := uuid.Scan(s)
	if err != nil {
		logger.Error("Не смогли спарсить UUID", zap.Error(err))
		return pgtype.UUID{}
	}
	return uuid
}
