package publishersRepository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
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

func (r *repository) Create(ctx context.Context, p *publisher.Publisher) (*publisher.Publisher, error) {
	q := lib.FormatQuery(`
		INSERT INTO publishers(name, country, city, point) 
		VALUES ($1, $2, $3, $4) 
		RETURNING publisher_id, add_date
	`)

	err := r.client.
		QueryRow(ctx, q, p.Name, p.Country, p.City, p.Point).
		Scan(&p.PublisherID, &p.AddDate)
	logger.Info(q)

	return p, lib.HandlePgErr(err)
}

func (r *repository) FindAll(ctx context.Context) (p []*publisher.Publisher, err error) {
	q := lib.FormatQuery(`
		SELECT 	p.name, p.publisher_id, p.add_date, p.country, p.city, p.point
		FROM publishers p
	`)

	rows, err := r.client.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	logger.Info(q)

	for rows.Next() {
		src := &publisher.Publisher{}
		if err = rows.Scan(&src.Name, &src.PublisherID, &src.AddDate, &src.Country, &src.City, &src.Point); err != nil {
			return nil, err
		}
		p = append(p, src)
	}
	return p, nil
}

func (r *repository) FindPublishersByName(ctx context.Context, name string) (p []*publisher.Publisher, err error) {
	q := lib.FormatQuery(`
		SELECT 	p.name, p.publisher_id, p.add_date, p.country, p.city, p.point
		FROM publishers p
		WHERE LOWER(p.name) LIKE LOWER('%'||$1||'%')
	`)

	rows, err := r.client.Query(ctx, q, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			logger.Error(pgErr.Message, zap.Error(pgErr))
			return nil, pgErr
		}
		return nil, err
	}

	logger.Info(q)

	for rows.Next() {
		src := &publisher.Publisher{}
		if err = rows.Scan(&src.Name, &src.PublisherID, &src.AddDate, &src.Country, &src.City, &src.Point); err != nil {
			return nil, err
		}
		p = append(p, src)
	}
	return p, nil
}

func (r *repository) FindPublishersByIDs(ctx context.Context, ids []pgtype.UUID) (p []*publisher.Publisher, err error) {
	q := lib.FormatQuery(`
		SELECT 	p.name, p.publisher_id, p.add_date, p.country, p.city, p.point
		FROM publishers p
		WHERE p.publisher_id = any($1)
	`)

	rows, err := r.client.Query(ctx, q, ids)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			logger.Error(pgErr.Message, zap.Error(pgErr))
			return nil, pgErr
		}
		return nil, err
	}

	logger.Info(q)

	for rows.Next() {
		src := &publisher.Publisher{}
		if err = rows.Scan(&src.Name, &src.PublisherID, &src.AddDate, &src.Country, &src.City, &src.Point); err != nil {
			return nil, err
		}
		p = append(p, src)
	}
	return p, nil
}

func (r *repository) Update(ctx context.Context, p *publisher.Publisher) error {
	q := lib.FormatQuery(`
		UPDATE publishers
		SET name = $1
		WHERE publisher_id = $2
	`)

	_, err := r.client.Query(ctx, q, p.Name, p.PublisherID)
	logger.Info(q)

	return lib.HandlePgErr(err)
}

func (r *repository) Delete(ctx context.Context, id string) error {
	begin, err := r.client.Begin(ctx)
	if err != nil {
		return lib.HandlePgErr(err)
	}

	// Удалить RSS
	q1 := lib.FormatQuery(`
		DELETE FROM sources_rss
		WHERE publisher_id = $1
	`)
	logger.Info(q1)
	if _, err = begin.Exec(ctx, q1, id); err != nil {
		return lib.HandlePgErr(err)
	}
	// Удалить источинк
	q2 := lib.FormatQuery(`
		DELETE FROM publishers
		WHERE publisher_id = $1
	`)
	logger.Info(q2)
	if _, err = begin.Exec(ctx, q2, id); err != nil {
		return lib.HandlePgErr(err)
	}

	if err = begin.Commit(ctx); err != nil {
		return lib.HandlePgErr(err)
	}

	return lib.HandlePgErr(err)
}
