package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"techno-forum/src/models"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ThreadRepository struct {
	dbpool *pgxpool.Pool
}

func NewThreadRepository(dbpool *pgxpool.Pool) *ThreadRepository {
	return &ThreadRepository{
		dbpool: dbpool,
	}
}

func (repo *ThreadRepository) GetBySlug(slug string) (*models.Thread, error) {
	var thread models.Thread
	var created time.Time
	err := repo.dbpool.QueryRow(context.Background(),
		`SELECT t.id, t.title, u.nickname, f.slug, f.id,
		t.message, t.votes_cnt, t.slug, t.created_at
		FROM Threads t 
		JOIN users u ON t.author_id = u.id
		JOIN forums f ON t.forum_id = f.id
		WHERE lower(t.slug) = lower($1)`, slug).
		Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.ForumId,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&created,
		)

	thread.Created = created.Format("2006-01-02T15:04:05.000Z")

	if err == nil {
		return &thread, nil
	}

	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return nil, err
}

func (repo *ThreadRepository) GetById(id string) (*models.Thread, error) {
	var thread models.Thread
	var created time.Time
	err := repo.dbpool.QueryRow(context.Background(),
		`SELECT t.id, t.title, u.nickname, f.slug, f.id,
		t.message, t.votes_cnt, t.slug, t.created_at
		FROM Threads t 
		JOIN users u ON t.author_id = u.id
		JOIN forums f ON t.forum_id = f.id
		WHERE t.id = $1`, id).
		Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.ForumId,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&created,
		)

	thread.Created = created.Format("2006-01-02T15:04:05.000Z")

	if err == nil {
		return &thread, nil
	}

	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return nil, err
}

func (repo *ThreadRepository) GetByForum(forumId int, since string, desc bool, limit int) ([]*models.Thread, error) {
	query := `SELECT t.id, t.title, u.nickname, f.slug, f.id,
	t.message, t.votes_cnt, t.slug, t.created_at::timestamptz
	FROM Threads t 
	JOIN users u ON t.author_id = u.id
	JOIN forums f ON t.forum_id = f.id
	WHERE f.id = $1`

	var rows pgx.Rows
	var err error

	if since != "" {
		if desc {
			query += ` AND t.created_at <= $2 ORDER BY t.created_at DESC LIMIT $3`
		} else {
			query += ` AND t.created_at >= $2 ORDER BY t.created_at LIMIT $3`
		}

		rows, err = repo.dbpool.Query(context.Background(),
			query,
			forumId,
			since,
			limit)
	} else {
		query += " ORDER BY t.created_at"
		if desc {
			query += " DESC"
		}
		query += " LIMIT $2"

		rows, err = repo.dbpool.Query(context.Background(),
			query,
			forumId,
			limit)
	}

	if err != nil {
		return nil, err
	}

	threads := []*models.Thread{}

	defer rows.Close()

	for rows.Next() {
		thread := models.Thread{}
		var created time.Time
		err = rows.Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.ForumId,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&created,
		)

		thread.Created = created.Format("2006-01-02T15:04:05.000Z")

		if err != nil {
			log.Fatal(err)
		}

		threads = append(threads, &thread)
	}

	return threads, nil
}

func (repo *ThreadRepository) Create(thread *models.Thread, author_id int, forum_id int) error {
	var err error

	if thread.Created != "" {
		timeParseLayout := "2006-01-02T15:04:05.000-07:00"
		time, t_err := time.Parse(timeParseLayout, thread.Created)
		thread.Created = time.UTC().Format("2006-01-02T15:04:05.000Z")

		fmt.Println("CREATED:", thread.Created)

		if t_err != nil {
			log.Fatal(err)
		}

		err = repo.dbpool.QueryRow(context.Background(),
			`INSERT INTO Threads (title, author_id, forum_id, message, created_at, slug) 
			values ($1, $2, $3, $4, $5, $6) RETURNING id`,
			thread.Title,
			author_id,
			forum_id,
			thread.Message,
			thread.Created,
			thread.Slug,
		).Scan(&thread.Id)
	} else {
		err = repo.dbpool.QueryRow(context.Background(),
			`INSERT INTO Threads (title, author_id, forum_id, message, slug) 
			values ($1, $2, $3, $4, $5) RETURNING id`,
			thread.Title,
			author_id,
			forum_id,
			thread.Message,
			thread.Slug,
		).Scan(&thread.Id)
	}

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return err
	}

	var created time.Time

	err = repo.dbpool.QueryRow(context.Background(),
		`SELECT t.id, t.title, u.nickname, f.slug,
			 t.message, t.votes_cnt, t.slug, t.created_at
	 FROM threads t JOIN users u ON t.author_id = u.id
					JOIN forums f ON t.forum_id  = f.id
	 WHERE lower(t.slug) = lower($1)`, thread.Slug).
		Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&created,
		)
	if err != nil {
		return err
	}

	thread.Created = created.Format("2006-01-02T15:04:05.000Z")

	return models.ErrAlreadyExists
}

func (repo *ThreadRepository) Update(thread *models.Thread) error {
	_, err := repo.dbpool.Exec(context.Background(),
		`UPDATE Threads SET 
						title = $1,
						message = $2
						WHERE id = $3`, thread.Title, thread.Message, thread.Id)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return models.ErrAlreadyExists
	}
	return err
}
