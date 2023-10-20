package repositories

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"golang.org/x/image/draw"
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

	createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
	if err != nil {
		panic(err)
	}
	createdAtStr := createdAt.Format("Jan 2, 2006")

	return &models.PostContent{
		Id:        post.Id,
		Title:     post.Title,
		Date:      createdAtStr,
		ImageMime: post.ImageMime,
		Image:     base64.StdEncoding.EncodeToString(post.Image),
		Content:   post.Content,
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
		postPreview := post.Content
		if len(postPreview) > 100 {
			postPreview = postPreview[:100] + "..."
		}

		createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			panic(err)
		}
		createdAtStr := createdAt.Format("Jan 2, 2006")

		posts[i] = models.PostStub{
			Id:        post.Id,
			Title:     post.Title,
			Date:      createdAtStr,
			ImageMime: post.ImageMime,
			Image:     base64.StdEncoding.EncodeToString(post.Thumbnail),
			Preview:   postPreview,
		}
	}

	return posts
}

// AddImage implements services.BlogPostService.
func (self *PostRepository) AddImage(
	ctx context.Context,
	id string,
	mimeType string,
	image []byte,
) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/blog/"+id+"/upload", "update"); err != nil {
		post, err := self.postDao.GetPost(ctx, id)
		if err == database.NotFound {
			return services.AccessDenied
		}

		if userId, ok := ctx.Value("user_id").(string); !ok || post.Author != userId {
			return services.AccessDenied
		}
	}

	if mimeType != "image/png" && mimeType != "image/jpeg" {
		return services.InvalidMimeType
	}

	var thumbnailBuffer bytes.Buffer
	if mimeType == "image/png" {
		resizePng(bytes.NewReader(image), &thumbnailBuffer)
	} else if mimeType == "image/jpeg" {
		resizeJpeg(bytes.NewReader(image), &thumbnailBuffer)
	}

	err := self.postDao.AddImage(ctx, id, mimeType, image, thumbnailBuffer.Bytes())
	if err != nil {
		panic(err)
	}

	return nil
}

const Width = 600

func resizeJpeg(r io.Reader, w io.Writer) {
	src, err := jpeg.Decode(r)
	if err != nil {
		panic(err)
	}

	scale := float64(Width) / float64(src.Bounds().Max.X)
	height := int(float64(src.Bounds().Max.Y) * scale)

	dst := image.NewRGBA(image.Rect(0, 0, Width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	jpeg.Encode(w, dst, nil)
}

func resizePng(r io.Reader, w io.Writer) {
	src, err := png.Decode(r)
	if err != nil {
		panic(err)
	}

	scale := float64(Width) / float64(src.Bounds().Max.X)
	height := int(float64(src.Bounds().Max.Y) * scale)

	dst := image.NewRGBA(image.Rect(0, 0, Width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	png.Encode(w, dst)
}

// var _ services.BlogPostService = (*PostRepository)(nil)
