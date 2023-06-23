package repository

import (
	"context"
	"errors"
	"fmt"
	"techno-forum/src/models"
	"techno-forum/src/utils"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	dbpool *pgxpool.Pool
}

func NewPostRepo(dbpool *pgxpool.Pool) *PostRepository {
	return &PostRepository{
		dbpool: dbpool,
	}
}

func getAuthorIds(tx pgx.Tx, posts []*models.Post) (map[string]int, error) {
	res := make(map[string]int, len(posts))

	for _, post := range posts {
		_, ok := res[post.Author]
		if ok {
			continue
		}

		var id int
		err := tx.QueryRow(context.Background(),
			"SELECT id FROM users WHERE lower(nickname) = lower($1)", post.Author).Scan(&id)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, models.ErrNotFound
			}
			return nil, err
		}

		res[post.Author] = id
	}

	return res, nil
}

func LinkUsersToForum(tx pgx.Tx, forumId int, ids map[string]int) error {
	query := `INSERT INTO ForumUserLinks(user_id, forum_id) VALUES `

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, forumId)

	i := 2
	for _, id := range ids {
		args = append(args, id)
		query += fmt.Sprintf("($%d, $1),", i)
		i++
	}

	query = query[:len(query)-1] + " ON CONFLICT DO NOTHING"

	_, err := tx.Exec(context.Background(), query, args...)
	return err
}

func (repo *PostRepository) AddPosts(thread *models.Thread, posts []*models.Post) error {
	return utils.MakeTx(repo.dbpool, func(tx pgx.Tx) error {
		ids, err := getAuthorIds(tx, posts)
		if err != nil {
			return err
		}

		createdAt := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		query := `INSERT INTO Posts(thread_id, author_id, parent_id, message, created_at) VALUES `

		args := make([]interface{}, 0, len(posts)*3+2)
		args = append(args, thread.Id, createdAt)

		i := 3
		for _, post := range posts {
			post.Thread = thread.Id
			post.Forum = thread.Forum
			post.Created = createdAt

			query += fmt.Sprintf("($1, $%d, $%d, $%d, $2),", i, i+1, i+2)
			args = append(args, ids[post.Author], post.Parent, post.Message)
			i += 3
		}

		query = query[:len(query)-1] + " RETURNING id"

		rows, err := repo.dbpool.Query(context.Background(), query, args...)

		if err != nil {
			return err
		}

		postIds, err := pgx.CollectRows(rows, pgx.RowTo[int64])

		fmt.Println(err)

		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				return err
			}

			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return models.ErrNoParent
			}

			if pgErr.Code == pgerrcode.IntegrityConstraintViolation {
				return models.ErrInvalidParent
			}
		}

		for i, id := range postIds {
			posts[i].Id = id
		}

		_, err = tx.Exec(context.Background(),
			`UPDATE Forums SET posts_cnt = posts_cnt + $1 WHERE id = $2`,
			len(posts), thread.ForumId,
		)
		if err != nil {
			return err
		}

		return LinkUsersToForum(tx, thread.ForumId, ids)
	})
}

func (repo *PostRepository) GetPost(id int64) (*models.Post, error) {
	post := &models.Post{Id: id}

	var created time.Time
	err := repo.dbpool.QueryRow(context.Background(),
		`SELECT u.nickname, p.message, p.edited, f.slug, p.parent_id, p.thread_id, p.created_at
		 FROM Posts p JOIN users u  ON u.id = p.author_id
		 			 JOIN threads t ON t.id = p.thread_id
					 JOIN forums f  ON f.id = t.forum_id
		 WHERE p.id = $1`, id).
		Scan(
			&post.Author,
			&post.Message,
			&post.IsEdited,
			&post.Forum,
			&post.Parent,
			&post.Thread,
			&created,
		)

	post.Created = created.Format("2006-01-02T15:04:05.000Z")

	fmt.Println(err)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, err
	}

	return post, nil
}

func (repo *PostRepository) Update(post *models.Post) error {
	previous, err := repo.GetPost(post.Id)
	if err != nil {
		return err
	}

	if post.Message == "" {
		post.Message = previous.Message
	}

	if post.Message != previous.Message {
		post.IsEdited = true
	}

	_, err = repo.dbpool.Exec(context.Background(),
		`UPDATE Posts SET message = $1, edited = $2 WHERE id = $3`,
		post.Message, post.IsEdited, post.Id)

	if err != nil {
		return err
	}

	post.Author = previous.Author
	post.Forum = previous.Forum
	post.Created = previous.Created
	post.Parent = previous.Parent
	post.Thread = previous.Thread

	return nil
}

func (repo *PostRepository) GetPostsFlat(params *models.PostListParams) ([]*models.Post, error) {

	fmt.Println("Params:", params)

	query := `SELECT p.id, u.nickname, p.message, p.edited,
					 p.parent_id, p.thread_id, p.created_at
				FROM Posts p JOIN users u  ON u.id = p.author_id
				WHERE p.thread_id = $1 `

	args := []interface{}{params.ThreadId}

	if params.Since > 0 {
		args = append(args, params.Since)
		if !params.Desc {
			query += "AND p.id > $2"
		} else {
			query += "AND p.id < $2"
		}
	}

	query += " ORDER BY p.id "

	if params.Desc {
		query += "DESC "
	}

	args = append(args, params.Limit)
	query += fmt.Sprintf("LIMIT $%d", len(args))

	rows, err := repo.dbpool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}

	var created time.Time

	posts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.Post, error) {
		post := &models.Post{}
		err := row.Scan(
			&post.Id,
			&post.Author,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&created,
		)
		post.Created = created.Format("2006-01-02T15:04:05.000Z")
		return post, err
	})

	fmt.Println("POSTS:", posts, err)

	if err != nil {
		if err == pgx.ErrNoRows {
			return posts, nil
		}
		return nil, err
	}

	return posts, nil
}

func (repo *PostRepository) GetPostsTree(params *models.PostListParams) ([]*models.Post, error) {
	query := `SELECT p.id, u.nickname, p.message, p.edited,
					 p.parent_id, p.thread_id, p.created_at
			  FROM Posts p JOIN users u  ON u.id = p.author_id
			  WHERE p.thread_id = $1 `

	args := []interface{}{params.ThreadId}

	if params.Since > 0 {
		args = append(args, params.Since)
		if !params.Desc {
			query += "AND p.path > (SELECT path FROM Posts WHERE id = $2)"
		} else {
			query += "AND p.path < (SELECT path FROM Posts WHERE id = $2)"
		}
	}
	query += " ORDER BY p.path"

	if params.Desc {
		query += " DESC"
	}

	args = append(args, params.Limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := repo.dbpool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}

	var created time.Time

	posts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.Post, error) {
		post := &models.Post{}
		err := row.Scan(
			&post.Id,
			&post.Author,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&created,
		)
		post.Created = created.Format("2006-01-02T15:04:05.000Z")
		return post, err
	})

	if err != nil {
		if err == pgx.ErrNoRows {
			return posts, nil
		}
		return nil, err
	}

	return posts, nil
}

func (repo *PostRepository) GetPostsParent(params *models.PostListParams) ([]*models.Post, error) {
	query := `WITH parents AS (
			  SELECT p.id, u.nickname, p.message, p.edited,
					 p.parent_id, p.thread_id, p.created_at,
					 p.path as path
			  FROM Posts p JOIN users u  ON u.id = p.author_id
			  WHERE p.thread_id = $1 AND p.id = p.path[1] `

	args := []interface{}{params.ThreadId}

	if params.Since > 0 {
		args = append(args, params.Since)
		if !params.Desc {
			query += "AND p.id > (SELECT path[1] FROM Posts WHERE id = $2)"
		} else {
			query += "AND p.id < (SELECT path[1] FROM Posts WHERE id = $2)"
		}
	}

	query += " ORDER BY p.id"
	if params.Desc {
		query += " DESC"
	}

	args = append(args, params.Limit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	query += `), final AS (
				SELECT p.id, u.nickname, p.message, p.edited,
					   p.parent_id, p.thread_id, p.created_at,
					   p.path as path
				FROM Posts p JOIN users u  ON u.id = p.author_id
					   		JOIN parents  ON parents.id = p.path[1]
		   		WHERE p.id != p.path[1]
		 		UNION ALL
		 		SELECT * FROM parents)
				SELECT id, nickname, message, edited,
				   parent_id, thread_id, created_at
				FROM final ORDER BY path[1]`

	if params.Desc {
		query += " DESC"
	}

	query += " NULLS FIRST, path[2:]"

	rows, err := repo.dbpool.Query(context.Background(), query, args...)

	fmt.Println("ERROR:", err)

	if err != nil {
		return nil, err
	}

	var created time.Time

	posts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.Post, error) {
		post := &models.Post{}
		err := row.Scan(
			&post.Id,
			&post.Author,
			&post.Message,
			&post.IsEdited,
			&post.Parent,
			&post.Thread,
			&created,
		)
		post.Created = created.Format("2006-01-02T15:04:05.000Z")
		return post, err
	})

	fmt.Println(err)

	if err != nil {
		if err == pgx.ErrNoRows {
			return posts, nil
		}
		return nil, err
	}

	return posts, nil
}
