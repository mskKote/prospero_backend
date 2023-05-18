package sourcesRepository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
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

func (r *repository) Create(ctx context.Context, s *source.RSS) (*source.RSS, error) {
	q := lib.FormatQuery(`
		INSERT INTO sources_rss(rss_url, publisher_id) 
		VALUES ($1, $2) 
		RETURNING rss_id
	`)

	err := r.client.
		QueryRow(ctx, q, s.RssURL, s.Publisher.PublisherID).
		Scan(&s.RssID)
	logger.Info(q)

	return s, lib.HandleErr(err)
}

func (r *repository) FindAll(ctx context.Context) (s []*source.RSS, err error) {
	q := lib.FormatQuery(`
		SELECT s.rss_id, s.rss_url, s.publisher_id
		FROM sources_rss s
	`)

	rows, err := r.client.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	logger.Info(q)

	for rows.Next() {
		src := &source.RSS{}
		if err = rows.Scan(&src.RssID, &src.RssURL, &src.Publisher.PublisherID); err != nil {
			return nil, err
		}
		s = append(s, src)
		logger.Info(lib.UuidToString(src.RssID))
	}
	return s, nil
}

func (r *repository) FindByPublisherName(ctx context.Context, name string) (s []*source.RSS, err error) {
	q := lib.FormatQuery(`
		SELECT 	s.rss_id, s.rss_url, 
				p.name, p.publisher_id, p.add_date, p.country, p.city, p.point
		FROM sources_rss s
			JOIN publishers p on p.publisher_id = s.publisher_id
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
		src := &source.RSS{}
		err = rows.Scan(
			&src.RssID, &src.RssURL,
			&src.Publisher.Name, &src.Publisher.PublisherID,
			&src.Publisher.AddDate, &src.Publisher.Country,
			&src.Publisher.City, &src.Publisher.Point)
		if err != nil {
			return nil, err
		}
		s = append(s, src)
	}
	return s, nil
}

func (r *repository) Update(ctx context.Context, s source.RSS) error {
	q := lib.FormatQuery(`
		UPDATE sources_rss
		SET rss_url = $1, publisher_id = $2
		WHERE rss_id = $3
	`)

	_, err := r.client.Query(ctx, q, s.RssURL, s.Publisher.PublisherID, s.RssID)
	logger.Info(q)

	return lib.HandleErr(err)
}

func (r *repository) Delete(ctx context.Context, id string) error {
	q := lib.FormatQuery(`
		DELETE FROM sources_rss
		WHERE rss_id = $1 
	`)

	_, err := r.client.Query(ctx, q, id)
	logger.Info(q)

	return lib.HandleErr(err)
}
