package repository

import (
	"context"
	"errors"
	"techno-forum/src/models"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ForumRepository struct {
	dbpool *pgxpool.Pool
}

func NewForumRepository(dbpool *pgxpool.Pool) *ForumRepository {
	return &ForumRepository{
		dbpool: dbpool,
	}
}

func (repo *ForumRepository) Create(forum *models.Forum, author_id int) error {
	_, err := repo.dbpool.Exec(context.Background(),
		"INSERT INTO Forums (title, slug, author_id) VALUES ($1, $2, $3)", forum.Title, forum.Slug, author_id)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return err
	}

	err = repo.dbpool.QueryRow(context.Background(),
		`SELECT f.id, u.nickname, f.title,
			   f.slug, f.posts_cnt, f.threads_cnt
		FROM Forums f 
		JOIN users u ON f.author_id = u.id
		WHERE lower(f.slug) = lower($1)`, forum.Slug).
		Scan(
			&forum.Id,
			&forum.Author,
			&forum.Title,
			&forum.Slug,
			&forum.Posts,
			&forum.Threads,
		)

	if err != nil {
		return err
	}

	return models.ErrAlreadyExists
}

func (repo *ForumRepository) Get(slug string) (*models.Forum, error) {
	forum := &models.Forum{}

	err := repo.dbpool.QueryRow(context.Background(),
		`SELECT f.id, u.nickname, f.title,
			f.slug, f.posts_cnt, f.threads_cnt
		FROM Forums f 
		JOIN users u ON f.author_id = u.id
		WHERE lower(f.slug) = lower($1)`, slug).
		Scan(
			&forum.Id,
			&forum.Author,
			&forum.Title,
			&forum.Slug,
			&forum.Posts,
			&forum.Threads,
		)

	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return forum, nil
}
