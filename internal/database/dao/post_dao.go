package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/database"
)

type PostDao struct {
	databaseProvider database.DatabaseProvider
}

func NewPostDao(databaseProvider database.DatabaseProvider) *PostDao {
	return &PostDao{databaseProvider: databaseProvider}
}

func (self *PostDao) CreatePost(
	ctx context.Context,
	title, content, author string,
) (string, error) {
	db := self.databaseProvider.Get()
	id := uuid.New().String()

	_, err := db.ExecContext(ctx, `
		INSERT INTO post (id, title, content, author)
		VALUES (?, ?, ?, ?)
	`, id, title, content, author)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (self *PostDao) GetPost(ctx context.Context, id string) (*database.Post, error) {
	db := self.databaseProvider.Get()

	var post database.Post
	err := db.GetContext(ctx, &post, `
		SELECT 
			id, title, content, author, image, image_mime, thumbnail, created_at, updated_at
		FROM post 
		WHERE id = ?
	`, id)
	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (self *PostDao) UpdatePost(
	ctx context.Context,
	id, title, content string,
) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		UPDATE post
		SET title = ?, content = ?
		WHERE id = ?
	`, title, content, id)
	if err == sql.ErrNoRows {
		return database.NotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (self *PostDao) DeletePost(ctx context.Context, id string) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM post
		WHERE id = ?
	`, id)
	if err == sql.ErrNoRows {
		return database.NotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (self *PostDao) ListPosts(ctx context.Context) ([]database.Post, error) {
	db := self.databaseProvider.Get()

	var posts []database.Post
	err := db.SelectContext(ctx, &posts, `
		SELECT 
			id, title, content, author, image, image_mime, thumbnail, created_at, updated_at
		FROM post
	`)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (self *PostDao) AddImage(ctx context.Context, id string, mimeType string, data []byte, thumbnail []byte) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		UPDATE post 
		SET image = ?, image_mime = ?, thumbnail = ?
		WHERE id = ?
	`, data, mimeType, thumbnail, id)

	if err != nil {
		return err
	}

	return nil
}
