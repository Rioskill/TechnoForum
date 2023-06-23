package usecase

import (
	"techno-forum/src/models"
	"techno-forum/src/repository"
)

type PostUseCase struct {
	PostRepo  *repository.PostRepository
	ForumRepo *repository.ForumRepository
}

func NewPostUseCase(posts *repository.PostRepository, forum *repository.ForumRepository) *PostUseCase {
	return &PostUseCase{
		PostRepo:  posts,
		ForumRepo: forum,
	}
}

func (usecase *PostUseCase) AddPosts(thread *models.Thread, posts []*models.Post) error {
	return usecase.PostRepo.AddPosts(thread, posts)
}

func (usecase *PostUseCase) GetPost(id int64) (*models.Post, error) {
	return usecase.PostRepo.GetPost(id)
}

func (usecase *PostUseCase) Update(post *models.Post) error {
	return usecase.PostRepo.Update(post)
}

func (u *PostUseCase) GetPosts(thread *models.Thread, params *models.PostListParams) ([]*models.Post, error) {
	var posts []*models.Post
	var err error

	switch params.Sort {
	case models.SortFlat:
		posts, err = u.PostRepo.GetPostsFlat(params)
	case models.SortTree:
		posts, err = u.PostRepo.GetPostsTree(params)
	case models.SortParent:
		posts, err = u.PostRepo.GetPostsParent(params)
	default:
		return nil, models.ErrInvalidArgument
	}

	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		post.Forum = thread.Forum
	}
	return posts, nil
}
