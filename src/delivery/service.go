package delivery

import (
	"encoding/json"
	"log"
	"net/http"
	"techno-forum/src/repository"
)

type ServiceDelivery struct {
	repo *repository.ServiceRepository
}

func NewServiceDelivery(repo *repository.ServiceRepository) *ServiceDelivery {
	return &ServiceDelivery{
		repo: repo,
	}
}

func (delivery *ServiceDelivery) Status(w http.ResponseWriter, r *http.Request) {
	info, err := delivery.repo.Status()

	if err != nil {
		log.Fatal(err)
	}

	res, err := json.Marshal(info)

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(200)
	status, err := w.Write(res)

	if err != nil {
		log.Fatal(status, err)
	}
}

func (delivery *ServiceDelivery) Clear(w http.ResponseWriter, r *http.Request) {
	err := delivery.repo.Clear()

	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(200)
}
