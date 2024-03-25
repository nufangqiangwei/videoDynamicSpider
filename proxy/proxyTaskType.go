package proxy

const (
	get  = "get"
	post = "post"
)

var VideoDetail = proxyMethod{
	Path:       "videoDetail",
	httpMethod: get,
	syncMethod: true,
}
var SyncVideoListDetail = proxyMethod{
	Path:       "SyncVideoListDetail",
	httpMethod: post,
	syncMethod: true,
}

var AuthorVideoList = proxyMethod{
	Path:       "authorVideoList",
	httpMethod: post,
	syncMethod: true,
}

var RecommendVideo = proxyMethod{
	Path:       "bilRecommendVideo",
	httpMethod: post,
	syncMethod: false,
}

var GetTaskStatus = proxyMethod{
	Path:       "getTaskStatus",
	httpMethod: get,
	syncMethod: false,
}
