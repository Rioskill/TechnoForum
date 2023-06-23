package repository

import (
	"context"
	"techno-forum/src/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceRepository struct {
	dbpool *pgxpool.Pool
}

func NewServiceRepo(dbpool *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{
		dbpool: dbpool,
	}
}

func (repo *ServiceRepository) Clear() error {
	_, err := repo.dbpool.Exec(context.Background(), "TRUNCATE users CASCADE")
	return err
}

func (repo *ServiceRepository) Status() (*models.ServiceInfo, error) {
	res := &models.ServiceInfo{}

	err := repo.dbpool.QueryRow(context.Background(), "SELECT count(*) FROM users").Scan(&res.User)

	if err != nil {
		return nil, err
	}

	err = repo.dbpool.QueryRow(context.Background(),
		`SELECT count(*), COALESCE(sum(threads_cnt), 0), COALESCE(sum(posts_cnt), 0) FROM forums`).
		Scan(&res.Forum, &res.Thread, &res.Post)
	if err != nil {
		return nil, err
	}

	return res, nil
}
