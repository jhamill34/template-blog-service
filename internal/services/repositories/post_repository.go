package repositories

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type PostRepository struct {
	postDao              *dao.PostDao
	accessControlService services.AccessControlService
}

func NewPostRepository(
	postDao *dao.PostDao,
	accessControlService services.AccessControlService,
) *PostRepository {
	return &PostRepository{
		postDao:              postDao,
		accessControlService: accessControlService,
	}
}

// CreatePost implements services.BlogPostService.
func (self *PostRepository) CreatePost(
	ctx context.Context,
	title string,
	content string,
	author string,
) (*models.PostStub, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/blog", "create"); err != nil {
		return nil, services.AccessDenied
	}

	postId, err := self.postDao.CreatePost(ctx, title, content, author)
	if err != nil {
		panic(err)
	}

	return &models.PostStub{
		Id:    postId,
		Title: title,
	}, nil
}

// GetPost implements services.BlogPostService.
func (self *PostRepository) GetPost(
	ctx context.Context,
	id string,
) (*models.PostContent, models.Notifier) {
	post, err := self.postDao.GetPost(ctx, id)
	if err == database.NotFound {
		return nil, services.PostNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.PostContent{
		Id:      post.Id,
		Title:   post.Title,
		Content: post.Content,
	}, nil
}

// UpdatePost implements services.BlogPostService.
func (self *PostRepository) UpdatePost(
	ctx context.Context,
	id string,
	title string,
	content string,
) (*models.PostStub, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/blog/"+id, "update"); err != nil {
		post, err := self.postDao.GetPost(ctx, id)
		if err == database.NotFound {
			return nil, services.AccessDenied
		}

		if userId, ok := ctx.Value("user_id").(string); !ok || post.Author != userId {
			return nil, services.AccessDenied
		}
	}

	err := self.postDao.UpdatePost(ctx, id, title, content)
	if err == database.NotFound {
		return nil, services.PostNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.PostStub{
		Id:    id,
		Title: title,
	}, nil
}

// DeletePost implements services.BlogPostService.
func (self *PostRepository) DeletePost(ctx context.Context, id string) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/blog/"+id, "delete"); err != nil {
		post, err := self.postDao.GetPost(ctx, id)
		if err == database.NotFound {
			return services.AccessDenied
		}

		if userId, ok := ctx.Value("user_id").(string); !ok || post.Author != userId {
			return services.AccessDenied
		}
	}

	err := self.postDao.DeletePost(ctx, id)
	if err == database.NotFound {
		return services.PostNotFound
	}

	if err != nil {
		panic(err)
	}

	return nil
}

// ListPosts implements services.BlogPostService.
func (self *PostRepository) ListPosts(ctx context.Context) []models.PostStub {
	data, err := self.postDao.ListPosts(ctx)

	if err != nil {
		panic(err)
	}

	posts := make([]models.PostStub, len(data))

	for i, post := range data {
		posts[i] = models.PostStub{
			Id:    post.Id,
			Title: post.Title,
		}
	}

	return posts
}

var _ services.BlogPostService = (*PostRepository)(nil)
