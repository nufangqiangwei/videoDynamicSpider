package proto

import "videoDynamicAcquisition/models"

func (a *AuthorInfoResponse) ToModel() *models.Author {
	author := models.Author{}
	author.WebSiteId = a.WebSiteId
	author.AuthorWebUid = a.Uid
	author.AuthorName = a.Name
	author.Avatar = a.Avatar
	author.AuthorDesc = a.Desc
	author.FollowNumber = a.FollowNumber
	return &author
}

func (t *TagInfoResponse) ToModel() *models.Tag {
	tag := models.Tag{}
	tag.Name = t.Name
	return &tag
}
