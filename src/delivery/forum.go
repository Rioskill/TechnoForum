package delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"techno-forum/src/models"
	"techno-forum/src/usecase"

	"github.com/go-chi/chi"
)

type ForumDelivery struct {
	usecase *usecase.ForumUseCase
}

func NewForumDelivery(usecase *usecase.ForumUseCase) *ForumDelivery {
	return &ForumDelivery{
		usecase: usecase,
	}
}

func (delivery *ForumDelivery) Create(w http.ResponseWriter, r *http.Request) {
	var forum models.Forum

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(string(reqBody))

	err = json.Unmarshal(reqBody, &forum)

	fmt.Println(forum)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = delivery.usecase.Create(&forum)

	if err == nil {
		res, err := json.Marshal(forum)

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
		status, err := w.Write([]byte(MakeErrorMsg("user not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrAlreadyExists {
		w.WriteHeader(409)
		res, err := json.Marshal(forum)

		if err != nil {
			log.Fatal(err)
		}

		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}

func (delivery *ForumDelivery) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	forum, err := delivery.usecase.Get(slug)

	if err == nil {
		res, err := json.Marshal(forum)

		if err != nil {
			log.Fatal(err)
		}

		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("forum not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}
	log.Fatal(err)
}
