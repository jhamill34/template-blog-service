package models

type PostStub struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	ImageMime string `json:"image_mime"`
	Image     string `json:"image"`
	Preview   string `json:"preview"`
}

type PostContent struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	ImageMime string `json:"image_mime"`
	Image     string `json:"image"`
	Content   string `json:"content"`
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
