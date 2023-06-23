package delivery

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"techno-forum/src/models"
	"techno-forum/src/repository"
	"techno-forum/src/usecase"

	"github.com/go-chi/chi"
)

type VoteDelivery struct {
	VoteRepo      *repository.VoteRepository
	UserRepo      *repository.UserRepository
	ThreadUseCase *usecase.ThreadUseCase
}

func NewVoteDelivery(VoteRepo *repository.VoteRepository,
	UserRepo *repository.UserRepository,
	ThreadUseCase *usecase.ThreadUseCase) *VoteDelivery {
	return &VoteDelivery{
		VoteRepo:      VoteRepo,
		UserRepo:      UserRepo,
		ThreadUseCase: ThreadUseCase,
	}
}

func (delivery *VoteDelivery) Vote(w http.ResponseWriter, r *http.Request) {
	slugOrId := chi.URLParam(r, "slugOrId")

	thread, err := delivery.ThreadUseCase.Get(slugOrId)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("thread not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	var voteRequest models.VoteRequest

	reqBody, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &voteRequest)

	if err != nil {
		log.Fatal(err)
	}

	user, err := delivery.UserRepo.GetByNickName(voteRequest.Nickname)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("user not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	vote := models.Vote{
		UserId:   user.Id,
		ThreadId: thread.Id,
		Value:    voteRequest.Voice,
	}

	err = delivery.VoteRepo.Vote(&vote)

	if err == models.ErrNotFound {
		w.WriteHeader(404)
		status, err := w.Write([]byte(MakeErrorMsg("not found")))

		if err != nil {
			log.Fatal(status, err)
		}
		return
	}

	thread, err = delivery.ThreadUseCase.Get(slugOrId)

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
