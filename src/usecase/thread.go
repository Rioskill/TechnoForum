package usecase

import (
	"fmt"
	"techno-forum/src/models"
	"techno-forum/src/repository"
	"techno-forum/src/utils"
)

type ThreadUseCase struct {
	ThreadRepo *repository.ThreadRepository
	UserRepo   *repository.UserRepository
	ForumRepo  *repository.ForumRepository
}

func NewThreadUseCase(thread *repository.ThreadRepository, user *repository.UserRepository, forum *repository.ForumRepository) *ThreadUseCase {
	return &ThreadUseCase{
		ThreadRepo: thread,
		UserRepo:   user,
		ForumRepo:  forum,
	}
}

func (usecase *ThreadUseCase) Create(thread *models.Thread, forumSlug string) error {
	user, err := usecase.UserRepo.GetByNickName(thread.Author)
	if err != nil {
		return err
	}

	forum, err := usecase.ForumRepo.Get(forumSlug)
	if err != nil {
		return err
	}

	thread.Author = user.Nickname
	thread.Forum = forum.Slug
	return usecase.ThreadRepo.Create(thread, user.Id, forum.Id)
}

func (usecase *ThreadUseCase) Get(slugOrId string) (*models.Thread, error) {
	if utils.IsNumeric(slugOrId) {
		return usecase.ThreadRepo.GetById(slugOrId)
	}
	return usecase.ThreadRepo.GetBySlug(slugOrId)
}

func (usecase *ThreadUseCase) GetByForum(forumSlug string, since string, desc bool, limit int) ([]*models.Thread, error) {
	forum, err := usecase.ForumRepo.Get(forumSlug)

	if err != nil {
		return nil, err
	}

	return usecase.ThreadRepo.GetByForum(forum.Id, since, desc, limit)
}

func (usecase *ThreadUseCase) Update(thread *models.Thread, slugOrId string) error {
	var foundThread *models.Thread
	var err error
	if utils.IsNumeric(slugOrId) {
		foundThread, err = usecase.ThreadRepo.GetById(slugOrId)
	} else {
		foundThread, err = usecase.ThreadRepo.GetBySlug(slugOrId)
	}

	if err != nil {
		return err
	}

	fmt.Println("FOUND:", foundThread)

	thread.Id = foundThread.Id
	thread.Slug = foundThread.Slug
	thread.Forum = foundThread.Forum
	thread.Author = foundThread.Author
	thread.Votes = foundThread.Votes
	thread.Created = foundThread.Created

	if thread.Title == "" {
		thread.Title = foundThread.Title
	}

	if thread.Message == "" {
		thread.Message = foundThread.Message
	}

	return usecase.ThreadRepo.Update(thread)
}
