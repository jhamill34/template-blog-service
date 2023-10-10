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

type ForwardError struct {
	Message string `json:"message"`
}

// Notify implements models.Notifier.
func (self *ForwardError) Notify() *Notification {
	return &Notification{
		Message: self.Message,
	}
}

