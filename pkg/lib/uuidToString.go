package lib

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func UuidToString(uuid pgtype.UUID) string {
	str, _ := uuid.Value()
	return str.(string)
	//return string(uuid.Bytes[:])
	//sb := strings.Builder{}
	//sb.Write(uuid.Bytes[:])
	//return sb.String()
}
