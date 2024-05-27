package pg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"yadro-microservices/internal/core/domain"
)

type ComicRepository struct {
	db *sql.DB
}

func NewComicRepository(db *sql.DB) *ComicRepository {
	return &ComicRepository{
		db: db,
	}
}

// Save saves comics to the database.
func (r *ComicRepository) Save(ctx context.Context, c domain.Comics) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error rolling back transaction: %v\n", err)
		}
	}(tx)

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO comics(id, img, keywords) VALUES($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for id, comic := range c {
		_, err = stmt.ExecContext(ctx, id, comic.Img, pq.Array(comic.Keywords))
		if err != nil {
			return fmt.Errorf("error executing statement: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetAll retrieves all comics from the database.
func (r *ComicRepository) GetAll(ctx context.Context) (domain.Comics, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, img, keywords FROM comics")
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	comics := make(domain.Comics)
	for rows.Next() {
		var id int
		var img string
		var keywords []string
		if err = rows.Scan(&id, &img, pq.Array(&keywords)); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		comics[id] = &domain.Comic{
			Img:      img,
			Keywords: keywords,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return comics, nil
}

// GetByID retrieves a comic by its ID from the database.
func (r *ComicRepository) GetByID(ctx context.Context, id int) (*domain.Comic, error) {
	row := r.db.QueryRowContext(ctx, "SELECT img, keywords FROM comics WHERE id = $1", id)

	var img string
	var keywords []string
	if err := row.Scan(&img, pq.Array(&keywords)); err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &domain.Comic{
		Img:      img,
		Keywords: keywords,
	}, nil
}

// GetAllIDs retrieves all existing comic IDs from the database.
func (r *ComicRepository) GetAllIDs(ctx context.Context) (map[int]bool, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM comics")
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	existingIDs := make(map[int]bool)
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		existingIDs[id] = true
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return existingIDs, nil
}

// GetTotalComics retrieves the total number of comics from the database.
func (r *ComicRepository) GetTotalComics(ctx context.Context) (int, error) {
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM comics")

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("error scanning row: %w", err)
	}

	return count, nil
}
