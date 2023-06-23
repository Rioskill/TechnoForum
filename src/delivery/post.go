package delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"techno-forum/src/models"
	"techno-forum/src/repository"
	"techno-forum/src/usecase"

	"github.com/go-chi/chi"
)

type PostDelivery struct {
	posts   *usecase.PostUseCase
	threads *usecase.ThreadUseCase
	forums  *usecase.ForumUseCase
	users   *repository.UserRepository
}

func NewPostDelivery(posts *usecase.PostUseCase,
	threads *usecase.ThreadUseCase,
	forums *usecase.ForumUseCase,
	users *repository.UserRepository) *PostDelivery {
	return &PostDelivery{
		posts:   posts,
		threads: threads,
		forums:  forums,
		users:   users,
	}
}

func (delivery *PostDelivery) Create(w http.ResponseWriter, r *http.Request) {
	var posts []*models.Post

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &posts)

	if err != nil {
		log.Fatal(err)
		return
	}

	slugOrId := chi.URLParam(r, "slugOrId")

	thread, err := delivery.threads.Get(slugOrId)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("thread not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	if len(posts) == 0 {
		w.WriteHeader(201)
		status, err := w.Write([]byte("[]"))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	err = delivery.posts.AddPosts(thread, posts)

	if err == nil {
		res, err := json.Marshal(posts)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(201)
		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrAlreadyExists {
		w.WriteHeader(409)
		status, err := w.Write([]byte(MakeErrorMsg("conflict")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrInvalidParent || err == models.ErrNoParent {
		w.WriteHeader(409)
		status, err := w.Write([]byte(MakeErrorMsg("conflict")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}

func (delivery *PostDelivery) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	post, err := delivery.posts.GetPost(id)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("post not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	fullPost := &models.PostFull{Post: post}

	relatedStr := r.URL.Query().Get("related")

	related := strings.Split(relatedStr, ",")

	for _, el := range related {
		switch el {
		case "user":
			user, err := delivery.users.GetByNickName(post.Author)
			if err != nil {
				// log.Fatal(err)
			}
			fullPost.Author = user
		case "forum":
			forum, err := delivery.forums.Get(post.Forum)
			if err != nil {
				// log.Fatal(err)
			}
			fullPost.Forum = forum
		case "thread":
			thread, err := delivery.threads.ThreadRepo.GetById(fmt.Sprint(post.Thread))
			if err != nil {
				// log.Fatal(err)
			}
			fullPost.Thread = thread
		}
	}

	res, err := json.Marshal(fullPost)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(200)
	status, err := w.Write(res)

	if err != nil {
		log.Fatal(status, err)
	}
}

func (delivery *PostDelivery) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		log.Fatal(err)
	}

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var post models.Post

	err = json.Unmarshal(reqBody, &post)
	post.Id = id

	err = delivery.posts.Update(&post)

	if err == nil {
		res, err := json.Marshal(post)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(200)
		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("post not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}
}

func (delivery *PostDelivery) GetByThread(w http.ResponseWriter, r *http.Request) {
	slugOrId := chi.URLParam(r, "slugOrId")

	thread, err := delivery.threads.Get(slugOrId)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("thread not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	var params models.PostListParams

	limitStr := r.URL.Query().Get("limit")
	sinceStr := r.URL.Query().Get("since")
	sort := r.URL.Query().Get("sort")
	descStr := r.URL.Query().Get("desc")

	var limit int
	if limitStr == "" {
		limit = 100
	} else {
		limit, err = strconv.Atoi(limitStr)

		if err != nil {
			log.Fatal(err)
		}
	}

	params.Limit = limit

	var since int
	if sinceStr == "" {
		since = 0
	} else {
		since, err = strconv.Atoi(sinceStr)

		if err != nil {
			log.Fatal(err)
		}
	}

	params.Since = since

	params.Desc = descStr == "true"

	switch sort {
	case "tree":
		params.Sort = models.SortTree
	case "parent_tree":
		params.Sort = models.SortParent
	default:
		params.Sort = models.SortFlat
	}

	params.ThreadId = thread.Id

	posts, err := delivery.posts.GetPosts(thread, &params)

	if err == nil {
		res, err := json.Marshal(posts)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(200)
		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}
}
