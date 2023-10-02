package models

type TemplateModel struct {
	Data interface{}
	Error interface{}
}

func NewTemplateData(data interface{}) TemplateModel {
	return TemplateModel{ Data: data }
}

func NewTemplateError(err interface{}) TemplateModel {
	return TemplateModel{ Error: err }
}

func NewTemplate(data interface{}, err interface{}) TemplateModel {
	return TemplateModel{ Data: data, Error: err }
}

func NewTemplateEmpty() TemplateModel {
	return TemplateModel{}
}

