package proto

import "strconv"

func (x *VideoInfoResponse) GetStructUniqueKey() string {
	return x.Uid
}
func (a *AuthorInfoResponse) GetStructUniqueKey() string {
	return a.Uid
}

func (t *TagInfoResponse) GetStructUniqueKey() string {
	return t.Name
}
func (c *CollectionInfo) GetStructUniqueKey() string {
	return strconv.FormatInt(c.CollectionId, 10)
}
