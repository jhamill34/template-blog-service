package models

type Envelope struct {
	Sender     string
	Recipients []string
	Data       []byte
}
