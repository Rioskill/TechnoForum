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
	PostsRepo := repository.NewPostRepo(dbpool)
	ServiceRepo := repository.NewServiceRepo(dbpool)
	VoteRepo := repository.NewVoteRepository(dbpool)

	ForumUseCase := usecase.NewForumUseCase(ForumRepo, UserRepo)
	ThreadUseCase := usecase.NewThreadUseCase(ThreadRepo, UserRepo, ForumRepo)
	PostsUseCase := usecase.NewPostUseCase(PostsRepo, ForumRepo)

	UserDelivery := delivery.NewUserDelivery(UserRepo, ForumRepo)
	ForumDelivery := delivery.NewForumDelivery(ForumUseCase)
	ThreadDelivery := delivery.NewThreadDelivery(ThreadUseCase)
	PostsDelivery := delivery.NewPostDelivery(PostsUseCase, ThreadUseCase, ForumUseCase, UserRepo)
	ServiceDelivery := delivery.NewServiceDelivery(ServiceRepo)
	VoteDelivery := delivery.NewVoteDelivery(VoteRepo, UserRepo, ThreadUseCase)

	r := chi.NewRouter()
	r.Use(ContentTypeSetter)

	r.Route("/api", func(r chi.Router) {
		r.Route("/forum", func(r chi.Router) {
			r.Post("/create", ForumDelivery.Create)
			r.Post("/{slug}/create", ThreadDelivery.Create)
			r.Get("/{slug}/details", ForumDelivery.Get)
			r.Get("/{slug}/threads", ThreadDelivery.GetByForum)
			r.Get("/{slug}/users", UserDelivery.GetByForum)
		})

		r.Route("/user", func(r chi.Router) {
			r.Post("/{nickname}/create", UserDelivery.Create)
			r.Get("/{nickname}/profile", UserDelivery.GetByNickName)
			r.Post("/{nickname}/profile", UserDelivery.Update)
		})

		r.Route("/thread", func(r chi.Router) {
			r.Get("/{slugOrId}/details", ThreadDelivery.Get)
			r.Post("/{slugOrId}/details", ThreadDelivery.Update)
			r.Post("/{slugOrId}/create", PostsDelivery.Create)
			r.Get("/{slugOrId}/posts", PostsDelivery.GetByThread)
			r.Post("/{slugOrId}/vote", VoteDelivery.Vote)
		})

		r.Route("/post", func(r chi.Router) {
			r.Get("/{id}/details", PostsDelivery.Get)
			r.Post("/{id}/details", PostsDelivery.Update)
		})

		r.Route("/service", func(r chi.Router) {
			r.Post("/clear", ServiceDelivery.Clear)
			r.Get("/status", ServiceDelivery.Status)
		})
	})

	log.Fatal(http.ListenAndServe(":5000", r))
}
