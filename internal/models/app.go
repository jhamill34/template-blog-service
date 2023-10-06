package models

type PostStub struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type PostContent struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
