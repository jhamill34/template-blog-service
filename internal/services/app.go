package services

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type BlogPostService interface {
	CreatePost(ctx context.Context, title, content, author string) (*models.PostStub, models.Notifier)
	GetPost(ctx context.Context, id string) (*models.PostContent, models.Notifier)
	UpdatePost(ctx context.Context, id, title, content string) (*models.PostStub, models.Notifier)
	DeletePost(ctx context.Context, id string) models.Notifier
	ListPosts(ctx context.Context) []models.PostStub
}
