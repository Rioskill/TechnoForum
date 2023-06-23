package usecase

import (
	"techno-forum/src/models"
	"techno-forum/src/repository"
)

type ForumUseCase struct {
	ForumRepo *repository.ForumRepository
	UserRepo  *repository.UserRepository
}

func NewForumUseCase(forum *repository.ForumRepository, user *repository.UserRepository) *ForumUseCase {
	return &ForumUseCase{
		ForumRepo: forum,
		UserRepo:  user,
	}
}

func (usecase *ForumUseCase) Create(forum *models.Forum) error {
	user, err := usecase.UserRepo.GetByNickName(forum.Author)
	if err != nil {
		return err
	}

	forum.Author = user.Nickname
	return usecase.ForumRepo.Create(forum, user.Id)
}

func (usecase *ForumUseCase) Get(slug string) (*models.Forum, error) {
	return usecase.ForumRepo.Get(slug)
}
