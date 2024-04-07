package main

import "videoDynamicAcquisition/models"

type webSite struct {
	data        []models.WebSite
	idToModel   map[int64]models.WebSite
	nameToModel map[string]models.WebSite
}

func (w *webSite) init() {
	w.data = make([]models.WebSite, 0)
	w.idToModel = make(map[int64]models.WebSite)
	w.nameToModel = make(map[string]models.WebSite)
	err := models.GormDB.Find(&w.data).Error
	if err != nil {
		panic(err)
	}
	for _, v := range w.data {
		w.idToModel[v.Id] = v
		w.nameToModel[v.WebName] = v
	}
}
func (w *webSite) getWebSiteById(webSiteId int64) *models.WebSite {
	web, ok := w.idToModel[webSiteId]
	if !ok {
		return nil
	}
	return &web
}

func (w *webSite) getWebSiteByName(webSiteName string) *models.WebSite {
	web, ok := w.nameToModel[webSiteName]
	if !ok {
		return nil
	}
	return &web
}
