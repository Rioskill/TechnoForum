package delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"techno-forum/src/models"
	"techno-forum/src/usecase"

	"github.com/go-chi/chi"
)

type ThreadDelivery struct {
	usecase *usecase.ThreadUseCase
}

func NewThreadDelivery(usecase *usecase.ThreadUseCase) *ThreadDelivery {
	return &ThreadDelivery{
		usecase: usecase,
	}
}

func (delivery *ThreadDelivery) Create(w http.ResponseWriter, r *http.Request) {
	var thread models.Thread

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &thread)

	if err != nil {
		log.Fatal(err)
	}

	forumSlug := chi.URLParam(r, "slug")

	err = delivery.usecase.Create(&thread, forumSlug)

	if err == nil {
		res, err := json.Marshal(thread)

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
		res, err := json.Marshal(thread)

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

	log.Fatal(err)
}

func (delivery *ThreadDelivery) Get(w http.ResponseWriter, r *http.Request) {
	SlugOrId := chi.URLParam(r, "slugOrId")

	thread, err := delivery.usecase.Get(SlugOrId)

	if err == nil {
		res, err := json.Marshal(thread)

		fmt.Println("thread:", thread)

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
		status, err := w.Write([]byte(MakeErrorMsg("thread not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}

func (delivery *ThreadDelivery) GetByForum(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	since := r.URL.Query().Get("since")
	descStr := r.URL.Query().Get("desc")
	limitStr := r.URL.Query().Get("limit")

	var limit int
	if limitStr == "" {
		limit = 100
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)

		if err != nil {
			log.Fatal(err)
		}
	}

	desc := descStr != "" && descStr != "false"

	thread, err := delivery.usecase.GetByForum(slug, since, desc, limit)

	if err == nil {
		res, err := json.Marshal(thread)

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
		status, err := w.Write([]byte(MakeErrorMsg("no threads found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}

func (delivery *ThreadDelivery) Update(w http.ResponseWriter, r *http.Request) {
	var thread models.Thread

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(string(reqBody))

	err = json.Unmarshal(reqBody, &thread)

	fmt.Println(thread)

	slugOrId := chi.URLParam(r, "slugOrId")

	err = delivery.usecase.Update(&thread, slugOrId)

	fmt.Println("THREAD", thread, err)

	if err == nil {
		res, err := json.Marshal(thread)

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
		status, err := w.Write([]byte(MakeErrorMsg("thread not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}
