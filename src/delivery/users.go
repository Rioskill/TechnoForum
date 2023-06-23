package delivery

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"techno-forum/src/models"
	"techno-forum/src/repository"
)

type UserDelivery struct {
	repo *repository.UserRepository
}

func NewUserDelivery(repo *repository.UserRepository) *UserDelivery {
	return &UserDelivery{
		repo: repo,
	}
}

func (delivery *UserDelivery) Create(w http.ResponseWriter, r *http.Request) {
	var p models.User

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.Nickname = chi.URLParam(r, "nickname")

	users, err := delivery.repo.Create(&p)

	if err == nil {
		res, err := json.Marshal(p)

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

	if users != nil {
		res, err := json.Marshal(users)

		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(409)
		status, err := w.Write(res)

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	log.Fatal(err)
}

func (delivery *UserDelivery) GetByNickName(w http.ResponseWriter, r *http.Request) {
	nickname := chi.URLParam(r, "nickname")

	user, err := delivery.repo.GetByNickName(nickname)

	if err == nil {
		res, err := json.Marshal(user)

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
		status, err := w.Write([]byte(MakeErrorMsg("user not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}
}

func (delivery *UserDelivery) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("update")
	var p models.User

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(string(reqBody))

	err = json.Unmarshal(reqBody, &p)

	fmt.Println(p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.Nickname = chi.URLParam(r, "nickname")

	err = delivery.repo.Update(&p)

	if err == nil {
		res, err := json.Marshal(p)

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
		status, err := w.Write([]byte(MakeErrorMsg("user not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err == models.ErrAlreadyExists {
		w.WriteHeader(409)
		status, err := w.Write([]byte(MakeErrorMsg("such user already exists")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}
}
