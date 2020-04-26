package types

import "log"

// прописываем стандартные routes
func (p *ProjectType) FillVueBaseRoutes() {
	if p.Vue.Routes == nil {
		p.Vue.Routes = [][]string{}
	}
	for _, d := range p.Docs {
		if len(d.Vue.RouteName) == 0 {
			log.Fatalf("'%s' missed RouteName", d.Name)
		}
		// индексы для возможных дублей роутов. Если найдутся аткие же, то перезаписываем
		indexRouteIndex := 0
		itemRouteIndex := 0
		for i, arr := range p.Vue.Routes {
			if len(arr) < 2 {
				log.Fatalf("'%s' in  project.Routes route: %v length < 2", arr)
			}
			if arr[0] == "/"+d.Name {
				indexRouteIndex = i
			}
			if arr[0] == "/"+d.Name+"/:id" {
				itemRouteIndex = i
			}
		}
		// route для index.vue
		if indexRouteIndex > 0 {
			p.Vue.Routes[indexRouteIndex] = []string{"/" + d.Name, d.Name + "/index.vue"}
		} else {
			p.Vue.Routes = append(p.Vue.Routes, []string{"/" + d.Name, d.Name + "/index.vue"})
		}
		// route для item.vue
		if itemRouteIndex > 0 {
			p.Vue.Routes[itemRouteIndex] = []string{"/" + d.Name + "/:id", d.Name + "/item.vue"}
		} else {
			p.Vue.Routes = append(p.Vue.Routes, []string{"/" + d.Name + "/:id", d.Name + "/item.vue"})
		}
	}
}