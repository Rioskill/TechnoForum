package repository

import (
	"context"
	"errors"
	"fmt"
	"techno-forum/src/models"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VoteRepository struct {
	dbpool *pgxpool.Pool
}

func NewVoteRepository(dbpool *pgxpool.Pool) *VoteRepository {
	return &VoteRepository{
		dbpool: dbpool,
	}
}

func (repo *VoteRepository) Vote(vote *models.Vote) error {
	_, err := repo.dbpool.Exec(context.Background(),
		`INSERT INTO vote(author_id, thread_id, value)
			VALUES($1, $2, $3) ON CONFLICT (author_id, thread_id) 
			DO UPDATE SET value = EXCLUDED.value`,
		vote.UserId, vote.ThreadId, vote.Value,
	)

	fmt.Println("ERR:", err)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return models.ErrNotFound
		}

		return err
	}

	return nil
}
