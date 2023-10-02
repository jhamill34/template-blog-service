package repositories

import (
	"html/template"
	"io"
	"path/filepath"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type TemplateRepository struct {
	common    *template.Template
	templates map[string]*template.Template
}

func NewTemplateRepository(components ...string) *TemplateRepository {
	var common *template.Template
	for ndx, component := range components {
		files, err := filepath.Glob(component)
		if err != nil {
			panic(err)
		}

		if ndx == 0 {
			common, err = template.ParseFiles(files...)
		} else {
			common, err = common.ParseFiles(files...)
		}

		if err != nil {
			panic(err)
		}
	}

	templates := make(map[string]*template.Template)

	return &TemplateRepository{
		templates: templates,
		common:    common,
	}
}

func (self *TemplateRepository) AddTemplates(filenames ...string) *TemplateRepository {
	for _, filename := range filenames {
		files, err := filepath.Glob(filename)

		if err != nil {
			panic(err)
		}

		for _, file := range files {
			common, err := self.common.Clone()

			if err != nil {
				panic(err)
			}

			template, err := common.ParseFiles(file)
			baseName := filepath.Base(file)

			self.templates[baseName] = template
		}
	}

	return self
}

// Render implements services.TemplateService.
func (self *TemplateRepository) Render(
	w io.Writer,
	name string,
	layout string,
	data models.TemplateModel,
) error {
	return self.templates[name].ExecuteTemplate(w, layout, data)
}

// var _ services.TemplateService = (*TemplateRepository)(nil)
