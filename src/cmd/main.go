package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"techno-forum/src/delivery"
	"techno-forum/src/repository"
	"techno-forum/src/usecase"
	"techno-forum/src/utils"

	"github.com/go-chi/chi"
)

func ContentTypeSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	dbpool, err := utils.InitPostgres()

	if err != nil {
		log.Fatal(err)
	}

	var greeting string

	err = dbpool.QueryRow(context.Background(), "select 'Hello, PostgeSQL!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return
	}

	fmt.Println(greeting)

	UserRepo := repository.NewUserRepo(dbpool)
	ForumRepo := repository.NewForumRepository(dbpool)
	ThreadRepo := repository.NewThreadRepository(dbpool)

	ForumUseCase := usecase.NewForumUseCase(ForumRepo, UserRepo)
	ThreadUseCase := usecase.NewThreadUseCase(ThreadRepo, UserRepo, ForumRepo)

	UserDelivery := delivery.NewUserDelivery(UserRepo)
	ForumDelivery := delivery.NewForumDelivery(ForumUseCase)
	ThreadDelivery := delivery.NewThreadDelivery(ThreadUseCase)

	r := chi.NewRouter()
	r.Use(ContentTypeSetter)

	r.Route("/api", func(r chi.Router) {
		r.Route("/forum", func(r chi.Router) {
			r.Post("/create", ForumDelivery.Create)
			r.Post("/{slug}/create", ThreadDelivery.Create)
			r.Get("/{slug}/details", ForumDelivery.Get)
			r.Get("/{slug}/threads", ThreadDelivery.GetByForum)
		})

		r.Route("/user", func(r chi.Router) {
			r.Post("/{nickname}/create", UserDelivery.Create)
			r.Get("/{nickname}/profile", UserDelivery.GetByNickName)
			r.Post("/{nickname}/profile", UserDelivery.Update)
		})

		r.Route("/thread", func(r chi.Router) {
			r.Get("/{slugOrId}/details", ThreadDelivery.Get)
			r.Get("/{slugOrId}/details", ThreadDelivery.Update)
		})
	})

	log.Fatal(http.ListenAndServe(":5000", r))
}
