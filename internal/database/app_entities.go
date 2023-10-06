package database

type Post struct {
	Id        string `db:"id"`
	Title     string `db:"title"`
	Content   string `db:"content"`
	Author    string `db:"author"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

