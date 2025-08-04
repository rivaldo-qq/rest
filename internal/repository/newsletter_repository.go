package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Dryluigi/go-grpc-ecommerce-be/internal/entity"
)

type INewsletterRepository interface {
	GetNewsletterByEmail(ctx context.Context, email string) (*entity.Newsletter, error)
	CreateNewNewsletter(ctx context.Context, newsletter *entity.Newsletter) error
}

type newsletterRepository struct {
	db *sql.DB
}

func (nr *newsletterRepository) GetNewsletterByEmail(ctx context.Context, email string) (*entity.Newsletter, error) {
	row := nr.db.QueryRowContext(
		ctx,
		"SELECT id FROM newsletter WHERE email = $1 AND is_deleted = false",
		email,
	)
	if row.Err() != nil {
		return nil, row.Err()
	}

	var newsletter entity.Newsletter
	err := row.Scan(
		&newsletter.Id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &newsletter, nil
}

func (nr *newsletterRepository) CreateNewNewsletter(ctx context.Context, newsletter *entity.Newsletter) error {
	_, err := nr.db.ExecContext(
		ctx,
		"INSERT INTO newsletter (id, full_name, email, created_at, created_by) VALUES ($1, $2, $3, $4, $5)",
		newsletter.Id,
		newsletter.FullName,
		newsletter.Email,
		newsletter.CreatedAt,
		newsletter.CreatedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewNewsletterRepository(db *sql.DB) INewsletterRepository {
	return &newsletterRepository{
		db: db,
	}
}
