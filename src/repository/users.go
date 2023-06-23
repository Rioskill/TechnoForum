package repository

import (
	"context"
	"errors"
	"fmt"

	"techno-forum/src/models"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbpool *pgxpool.Pool
}

func NewUserRepo(dbpool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		dbpool: dbpool,
	}
}

func (repo *UserRepository) Create(profile *models.User) ([]*models.User, error) {
	_, err := repo.dbpool.Exec(context.Background(),
		"INSERT INTO Users (nickname, fullname, email, about) values ($1, $2, $3, $4)",
		profile.Nickname,
		profile.Fullname,
		profile.Email,
		profile.About,
	)

	if err == nil {
		return nil, nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
		return nil, err
	}

	rows, err := repo.dbpool.Query(context.Background(),
		`SELECT id, nickname, fullname, about, email
		 FROM Users WHERE lower(email) = lower($1) OR lower(nickname) = lower($2)`,
		profile.Email, profile.Nickname)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*models.User{}

	for rows.Next() {
		userProfile := models.User{}
		err = rows.Scan(
			&userProfile.Id,
			&userProfile.Nickname,
			&userProfile.Fullname,
			&userProfile.About,
			&userProfile.Email,
		)

		users = append(users, &userProfile)
	}

	return users, models.ErrAlreadyExists
}

func (repo *UserRepository) GetByNickName(nickname string) (*models.User, error) {
	res := &models.User{}

	err := repo.dbpool.QueryRow(context.Background(),
		`SELECT id, nickname, fullname, about, email
		 FROM Users WHERE lower(nickname) = lower($1)`, nickname).
		Scan(&res.Id,
			&res.Nickname,
			&res.Fullname,
			&res.About,
			&res.Email)

	if err == pgx.ErrNoRows {
		return nil, models.ErrNotFound
	}

	return res, nil
}

func (repo *UserRepository) GetByForum(forumId int, limit int, since string, desc bool) ([]*models.User, error) {
	query := `SELECT u.id, u.nickname, u.fullname, u.about, u.email
				FROM users u JOIN ForumUserLinks uf ON u.id = uf.user_id
							JOIN forums f ON f.id = uf.forum_id
				WHERE f.id = $1 `

	args := []interface{}{forumId}

	if since != "" {
		query += "AND lower(u.nickname) "
		args = append(args, since)

		if !desc {
			query += "> lower($2)"
		} else {
			query += "< lower($2)"
		}
	}

	query += " ORDER BY lower(u.nickname)"
	if desc {
		query += " DESC"
	}

	args = append(args, limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := repo.dbpool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.User, error) {
		var user models.User
		err := row.Scan(
			&user.Id,
			&user.Nickname,
			&user.Fullname,
			&user.About,
			&user.Email,
		)
		return &user, err
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return users, nil
		}
		return nil, err
	}

	return users, nil
}

func (repo *UserRepository) Update(profile *models.User) error {
	user, err := repo.GetByNickName(profile.Nickname)

	if err != nil {
		return err
	}

	if profile.Nickname == "" {
		profile.Nickname = user.Nickname
	}

	if profile.Fullname == "" {
		profile.Fullname = user.Fullname
	}

	if profile.About == "" {
		profile.About = user.About
	}

	if profile.Email == "" {
		profile.Email = user.Email
	}

	_, err = repo.dbpool.Exec(context.Background(),
		`UPDATE Users SET 
						nickname = $1,
						fullname = $2,
						about = $3,
						email = $4
						WHERE id = $5`, profile.Nickname, profile.Fullname, profile.About, profile.Email, user.Id)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return models.ErrAlreadyExists
	}
	return err
}
