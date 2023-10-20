package database

type Post struct {
	Id        string `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Author    string `db:"author"`
	Image     []byte `db:"image"`
	Thumbnail []byte `db:"thumbnail"`
	ImageMime *string `db:"image_mime"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

