package lib

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

//pgtype.UUID{Bytes: [16]byte([]byte(dto.PublisherID))}

func StringToUUID(s string) pgtype.UUID {
	//return pgtype.UUID{Bytes: [16]byte([]byte(s))}
	uuid := pgtype.UUID{}
	err := uuid.Scan(s)
	if err != nil {
		logger.Error("Не смогли спарсить UUID", zap.Error(err))
		return pgtype.UUID{}
	}
	return uuid
}
