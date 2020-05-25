package types

import (
	"fmt"
	"github.com/pepelazz/projectGenerator/utils"
	"github.com/serenize/snaker"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type (
	ProjectType struct {
		Name     string
		Docs     []DocType
		DistPath string
		Config   ProjectConfig
		Vue      ProjectVue
		Sql      ProjectSql
	}
	ProjectConfig struct {
		Logo             string
		LocalProjectPath string
		Postgres         PostrgesConfig
		WebServer        WebServerConfig
		Email            EmailConfig
		DevMode          DevModeConfig
		Vue              VueConfig
	}
	PostrgesConfig struct {
		DbName   string
		Port     int64
		Password string
		TimeZone string // время для сервера default 'Europe/Moscow' (например 'Asia/Novosibirsk')
	}
	WebServerConfig struct {
		Port     int64
		Url      string
		Path     string
		Ip       string
		Username string // root или ...
	}
	DevModeConfig struct {
		IsDocker bool
	}
	EmailConfig struct {
		Sender     string
		Password   string
		Host       string
		Port       int64
		SenderName string
	}
	ProjectVue struct {
		UiAppName string
		Routes    [][]string
		Menu      []VueMenu
	}
	VueMenu struct {
		DocName  string // если указано docName, то url и иконка копируются из описания документа
		Icon     string
		Text     string
		Url      string
		IsFolder bool
		LinkList []VueMenu
		Roles    []string
	}

	VueConfig struct {
		DadataToken string
	}

	ProjectSql struct {
		Methods map[string][]DocSqlMethod // имя документа и список методов. Например "task": []{"task_by_deal"}
	}
)

func (p *ProjectType) GetDocByName(docName string) *DocType {
	for _, d := range p.Docs {
		if d.Name == docName {
			return &d
		}
	}
	return nil
}

// заполняем поля темплейтов - из короткой формы записи в полную
func (p *ProjectType) FillDocTemplatesFields() {
	for i, d := range p.Docs {
		if d.Templates == nil {
			d.Templates = map[string]*DocTemplate{}
		}
		for tName, t := range d.Templates {
			// прописываем полный путь к файлу шаблона
			if len(t.Source) == 0 {
				// учитывааем что возможен префикс, если папка с документом вложена в другую папку
				pathPrefix := ""
				if len(d.PathPrefix) > 0 {
					pathPrefix = d.PathPrefix + "/"
				}
				t.Source = fmt.Sprintf("%s%s/tmpl/%s", pathPrefix, snaker.SnakeToCamelLower(d.Name), tName)
			}
			// если не указан конечный путь, то формируем его исходя из ключа шаблона (например webClient_comp_...)
			if len(t.DistPath) == 0 {
				params := map[string]string{}
				if len(d.Vue.Path) > 0 {
					params["doc.Vue.Path"] = d.Vue.Path
				}
				distPath, distFilename := utils.ParseDocTemplateFilename(d.Name, tName, p.DistPath, i, params)
				t.DistFilename = distFilename
				t.DistPath = distPath
			}
		}
		p.Docs[i] = d
	}
}

// заполняем незаполненные поля для Vue
func (p *ProjectType) FillVueFlds() {
	for i, d := range p.Docs {
		for j, fld := range d.Flds {
			// если NameRu не заполнено, то копируем из fld
			if len(fld.Vue.NameRu) == 0 {
				p.Docs[i].Flds[j].Vue.NameRu = fld.NameRu
			}
			// заполняем IsRequired
			if fld.Sql.IsRequired {
				p.Docs[i].Flds[j].Vue.IsRequired = fld.Sql.IsRequired
			}
			// заполняем незаполненные поля в extension
			for k, _ := range fld.Vue.Ext {
				// если в параметрах есть pathUrl и поле является Ref, это значит надо заполнить route к доккументу, на который идет ссылка + ссылка на аватарку
				if k == "pathUrl" && len(fld.Sql.Ref)>0 {
					for _, dRef := range p.Docs {
						if dRef.Name == fld.Sql.Ref {
							fld.Vue.Ext["pathUrl"] = "/" + dRef.Vue.RouteName
							fld.Vue.Ext["avatar"] = dRef.Vue.MenuIcon
						}
					}
				}
			}
		}
	}
}

// заполняем боковое меню для Vue
func (p *ProjectType) FillSideMenu() {
	if (p.Vue.Menu == nil) {
		log.Fatalf("ProjectType.FillSideMenu p.Vue.Menu == nil")
	}
	for i, v := range p.Vue.Menu {
		if len(v.DocName) > 0 {
			d := p.GetDocByName(v.DocName)
			if d == nil {
				log.Fatalf("ProjectType.FillSideMenu p.GetDocByName doc '%s' not found", v.DocName)
			}
			if len(v.Icon) == 0 {
				p.Vue.Menu[i].Icon = d.Vue.MenuIcon
			}
			if len(v.Url) == 0 {
				p.Vue.Menu[i].Url = d.Vue.RouteName
			}
			if len(v.Text) == 0 {
				// если есть локализованное название для списка, то используем его (там множественное число). Если нет, то название документа
				if title, ok := d.Vue.I18n["listTitle"]; ok {
					p.Vue.Menu[i].Text = title
				} else {
					p.Vue.Menu[i].Text = utils.UpperCaseFirst(d.NameRu)
				}
			}
			if len(v.Roles) == 0 {
				p.Vue.Menu[i].Roles = d.Vue.Roles
			}
		}
		if v.IsFolder {
			for j, v1 := range v.LinkList {
				if len(v1.DocName) > 0 {
					d := p.GetDocByName(v1.DocName)
					if d == nil {
						log.Fatalf("ProjectType.FillSideMenu p.GetDocByName doc '%s' not found", v1.DocName)
					}
					if len(v1.Icon) == 0 {
						p.Vue.Menu[i].LinkList[j].Icon = d.Vue.MenuIcon
					}
					if len(v1.Url) == 0 {
						p.Vue.Menu[i].LinkList[j].Url = d.Vue.RouteName
					}
					if len(v1.Text) == 0 {
						// если есть локализованное название для списка, то используем его (там множественное число). Если нет, то название документа
						if title, ok := d.Vue.I18n["listTitle"]; ok {
							p.Vue.Menu[i].LinkList[j].Text = title
						} else {
							p.Vue.Menu[i].LinkList[j].Text = utils.UpperCaseFirst(d.NameRu)
						}
					}
					if len(v1.Roles) == 0 {
						p.Vue.Menu[i].LinkList[j].Roles = d.Vue.Roles
					}
				}
			}
		}
	}
}

// генерим сетку для Vue
func (p *ProjectType) GenerateGrid() {
	for i, d := range p.Docs {
		d.Vue.Grid = makeGrid(d)
		p.Docs[i] = d
	}
}

// если не указан путь к локальному проекту, то вычисляем его автоматически
func (p *ProjectType) FillLocalPath() string {
	if len(p.Config.LocalProjectPath) == 0 {
		// путь к локальной директории
		path, _ := filepath.Abs("./")
		// находим gopath
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		// убираем из начала пути gopath
		path = strings.TrimPrefix(path, gopath)
		// приводим разделитель пути к unix стилю
		path = strings.Replace(path, string(os.PathSeparator), "/", -1)
		// убираем из начала еще src
		path = strings.TrimPrefix(path, "/src/")
		// убираем из конца projectTemplate и добавляем src
		path = strings.TrimSuffix(path, "/projectTemplate") + "/src"
		p.Config.LocalProjectPath = path
	}
	return p.Config.LocalProjectPath
}
