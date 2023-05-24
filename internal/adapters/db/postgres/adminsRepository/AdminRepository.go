package adminsRepository

import (
	"context"
	"fmt"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
	"github.com/mskKote/prospero_backend/pkg/client/postgres"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
)

var logger = logging.GetLogger().With(zap.String("prefix", "[POSTGRES]"))

type repository struct {
	client postgres.Client
}

func New(client postgres.Client) IRepository {
	return &repository{client}
}

func (r *repository) FindAdminByName(ctx context.Context, name string) (*admin.Admin, error) {
	a := &admin.Admin{Name: name}
	q := lib.FormatQuery(`
		SELECT user_id, password
		FROM admins
		WHERE name=$1
	`)

	err := r.client.
		QueryRow(ctx, q, a.Name).
		Scan(&a.UserID, &a.Password)

	logger.Info(fmt.Sprintf("Запрашиваем админа %s", name))

	return a, lib.HandlePgErr(err)
}

func (r *repository) Create(ctx context.Context, a *admin.Admin) error {
	q := lib.FormatQuery(`
		INSERT INTO admins(name, password) 
		VALUES ($1, $2)
		RETURNING user_id
	`)

	err := r.client.QueryRow(ctx, q, a.Name, a.Password).Scan(&a.UserID)
	logger.Info(q)

	return lib.HandlePgErr(err)
}
