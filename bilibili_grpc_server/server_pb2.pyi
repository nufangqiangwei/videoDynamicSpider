from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class userInfo(_message.Message):
    __slots__ = ("cookies", "lastHistoryTime", "lastUpdateTime")
    class CookiesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    COOKIES_FIELD_NUMBER: _ClassVar[int]
    LASTHISTORYTIME_FIELD_NUMBER: _ClassVar[int]
    LASTUPDATETIME_FIELD_NUMBER: _ClassVar[int]
    cookies: _containers.ScalarMap[str, str]
    lastHistoryTime: str
    lastUpdateTime: str
    def __init__(self, cookies: _Optional[_Mapping[str, str]] = ..., lastHistoryTime: _Optional[str] = ..., lastUpdateTime: _Optional[str] = ...) -> None: ...

class videoInfoResponse(_message.Message):
    __slots__ = ("title", "desc", "cover", "uid", "duration", "updateTime", "collectTime", "tags", "classify", "viewNumber", "danmaku", "reply", "favorite", "coin", "share", "nowRank", "hisRank", "like", "dislike", "evaluation", "authors", "viewInfo", "webSiteName", "errorCode", "errorMsg", "requestUserName", "requestUserId", "webSiteId")
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESC_FIELD_NUMBER: _ClassVar[int]
    COVER_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    UPDATETIME_FIELD_NUMBER: _ClassVar[int]
    COLLECTTIME_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    CLASSIFY_FIELD_NUMBER: _ClassVar[int]
    VIEWNUMBER_FIELD_NUMBER: _ClassVar[int]
    DANMAKU_FIELD_NUMBER: _ClassVar[int]
    REPLY_FIELD_NUMBER: _ClassVar[int]
    FAVORITE_FIELD_NUMBER: _ClassVar[int]
    COIN_FIELD_NUMBER: _ClassVar[int]
    SHARE_FIELD_NUMBER: _ClassVar[int]
    NOWRANK_FIELD_NUMBER: _ClassVar[int]
    HISRANK_FIELD_NUMBER: _ClassVar[int]
    LIKE_FIELD_NUMBER: _ClassVar[int]
    DISLIKE_FIELD_NUMBER: _ClassVar[int]
    EVALUATION_FIELD_NUMBER: _ClassVar[int]
    AUTHORS_FIELD_NUMBER: _ClassVar[int]
    VIEWINFO_FIELD_NUMBER: _ClassVar[int]
    WEBSITENAME_FIELD_NUMBER: _ClassVar[int]
    ERRORCODE_FIELD_NUMBER: _ClassVar[int]
    ERRORMSG_FIELD_NUMBER: _ClassVar[int]
    REQUESTUSERNAME_FIELD_NUMBER: _ClassVar[int]
    REQUESTUSERID_FIELD_NUMBER: _ClassVar[int]
    WEBSITEID_FIELD_NUMBER: _ClassVar[int]
    title: str
    desc: str
    cover: str
    uid: str
    duration: int
    updateTime: int
    collectTime: int
    tags: _containers.RepeatedCompositeFieldContainer[tagInfoResponse]
    classify: _containers.RepeatedCompositeFieldContainer[classifyInfoResponse]
    viewNumber: int
    danmaku: int
    reply: int
    favorite: int
    coin: int
    share: int
    nowRank: int
    hisRank: int
    like: int
    dislike: int
    evaluation: str
    authors: _containers.RepeatedCompositeFieldContainer[AuthorInfoResponse]
    viewInfo: viewInfoResponse
    webSiteName: str
    errorCode: int
    errorMsg: str
    requestUserName: str
    requestUserId: int
    webSiteId: int
    def __init__(self, title: _Optional[str] = ..., desc: _Optional[str] = ..., cover: _Optional[str] = ..., uid: _Optional[str] = ..., duration: _Optional[int] = ..., updateTime: _Optional[int] = ..., collectTime: _Optional[int] = ..., tags: _Optional[_Iterable[_Union[tagInfoResponse, _Mapping]]] = ..., classify: _Optional[_Iterable[_Union[classifyInfoResponse, _Mapping]]] = ..., viewNumber: _Optional[int] = ..., danmaku: _Optional[int] = ..., reply: _Optional[int] = ..., favorite: _Optional[int] = ..., coin: _Optional[int] = ..., share: _Optional[int] = ..., nowRank: _Optional[int] = ..., hisRank: _Optional[int] = ..., like: _Optional[int] = ..., dislike: _Optional[int] = ..., evaluation: _Optional[str] = ..., authors: _Optional[_Iterable[_Union[AuthorInfoResponse, _Mapping]]] = ..., viewInfo: _Optional[_Union[viewInfoResponse, _Mapping]] = ..., webSiteName: _Optional[str] = ..., errorCode: _Optional[int] = ..., errorMsg: _Optional[str] = ..., requestUserName: _Optional[str] = ..., requestUserId: _Optional[int] = ..., webSiteId: _Optional[int] = ...) -> None: ...

class AuthorInfoResponse(_message.Message):
    __slots__ = ("author", "name", "avatar", "desc", "uid", "followNumber", "followTime", "webSiteName", "errorCode", "errorMsg", "requestUserName", "requestUserId", "webSiteId")
    AUTHOR_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    AVATAR_FIELD_NUMBER: _ClassVar[int]
    DESC_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    FOLLOWNUMBER_FIELD_NUMBER: _ClassVar[int]
    FOLLOWTIME_FIELD_NUMBER: _ClassVar[int]
    WEBSITENAME_FIELD_NUMBER: _ClassVar[int]
    ERRORCODE_FIELD_NUMBER: _ClassVar[int]
    ERRORMSG_FIELD_NUMBER: _ClassVar[int]
    REQUESTUSERNAME_FIELD_NUMBER: _ClassVar[int]
    REQUESTUSERID_FIELD_NUMBER: _ClassVar[int]
    WEBSITEID_FIELD_NUMBER: _ClassVar[int]
    author: str
    name: str
    avatar: str
    desc: str
    uid: str
    followNumber: int
    followTime: int
    webSiteName: str
    errorCode: int
    errorMsg: str
    requestUserName: str
    requestUserId: int
    webSiteId: int
    def __init__(self, author: _Optional[str] = ..., name: _Optional[str] = ..., avatar: _Optional[str] = ..., desc: _Optional[str] = ..., uid: _Optional[str] = ..., followNumber: _Optional[int] = ..., followTime: _Optional[int] = ..., webSiteName: _Optional[str] = ..., errorCode: _Optional[int] = ..., errorMsg: _Optional[str] = ..., requestUserName: _Optional[str] = ..., requestUserId: _Optional[int] = ..., webSiteId: _Optional[int] = ...) -> None: ...

class tagInfoResponse(_message.Message):
    __slots__ = ("name", "id", "errorCode", "errorMsg")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    ERRORCODE_FIELD_NUMBER: _ClassVar[int]
    ERRORMSG_FIELD_NUMBER: _ClassVar[int]
    name: str
    id: int
    errorCode: int
    errorMsg: str
    def __init__(self, name: _Optional[str] = ..., id: _Optional[int] = ..., errorCode: _Optional[int] = ..., errorMsg: _Optional[str] = ...) -> None: ...

class classifyInfoResponse(_message.Message):
    __slots__ = ("name", "id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    id: int
    def __init__(self, name: _Optional[str] = ..., id: _Optional[int] = ...) -> None: ...

class viewInfoResponse(_message.Message):
    __slots__ = ("viewTime", "viewDuration")
    VIEWTIME_FIELD_NUMBER: _ClassVar[int]
    VIEWDURATION_FIELD_NUMBER: _ClassVar[int]
    viewTime: int
    viewDuration: int
    def __init__(self, viewTime: _Optional[int] = ..., viewDuration: _Optional[int] = ...) -> None: ...

class collectionInfoResponse(_message.Message):
    __slots__ = ("name", "uid", "author", "creatTime", "appendTime", "errorCode", "errorMsg")
    NAME_FIELD_NUMBER: _ClassVar[int]
    UID_FIELD_NUMBER: _ClassVar[int]
    AUTHOR_FIELD_NUMBER: _ClassVar[int]
    CREATTIME_FIELD_NUMBER: _ClassVar[int]
    APPENDTIME_FIELD_NUMBER: _ClassVar[int]
    ERRORCODE_FIELD_NUMBER: _ClassVar[int]
    ERRORMSG_FIELD_NUMBER: _ClassVar[int]
    name: str
    uid: str
    author: AuthorInfoResponse
    creatTime: int
    appendTime: int
    errorCode: int
    errorMsg: str
    def __init__(self, name: _Optional[str] = ..., uid: _Optional[str] = ..., author: _Optional[_Union[AuthorInfoResponse, _Mapping]] = ..., creatTime: _Optional[int] = ..., appendTime: _Optional[int] = ..., errorCode: _Optional[int] = ..., errorMsg: _Optional[str] = ...) -> None: ...

class collectionInfo(_message.Message):
    __slots__ = ("collectionType", "collectionId", "collectionVideoCount", "lastAppendTime", "name", "upperName", "upperUid", "video")
    COLLECTIONTYPE_FIELD_NUMBER: _ClassVar[int]
    COLLECTIONID_FIELD_NUMBER: _ClassVar[int]
    COLLECTIONVIDEOCOUNT_FIELD_NUMBER: _ClassVar[int]
    LASTAPPENDTIME_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    UPPERNAME_FIELD_NUMBER: _ClassVar[int]
    UPPERUID_FIELD_NUMBER: _ClassVar[int]
    VIDEO_FIELD_NUMBER: _ClassVar[int]
    collectionType: str
    collectionId: str
    collectionVideoCount: int
    lastAppendTime: int
    name: str
    upperName: str
    upperUid: str
    video: _containers.RepeatedCompositeFieldContainer[videoInfoResponse]
    def __init__(self, collectionType: _Optional[str] = ..., collectionId: _Optional[str] = ..., collectionVideoCount: _Optional[int] = ..., lastAppendTime: _Optional[int] = ..., name: _Optional[str] = ..., upperName: _Optional[str] = ..., upperUid: _Optional[str] = ..., video: _Optional[_Iterable[_Union[videoInfoResponse, _Mapping]]] = ...) -> None: ...

class collectionInfoRequest(_message.Message):
    __slots__ = ("user", "collection")
    USER_FIELD_NUMBER: _ClassVar[int]
    COLLECTION_FIELD_NUMBER: _ClassVar[int]
    user: userInfo
    collection: _containers.RepeatedCompositeFieldContainer[collectionInfo]
    def __init__(self, user: _Optional[_Union[userInfo, _Mapping]] = ..., collection: _Optional[_Iterable[_Union[collectionInfo, _Mapping]]] = ...) -> None: ...

class getVideoListRequest(_message.Message):
    __slots__ = ("userInfo", "videoIdList")
    USERINFO_FIELD_NUMBER: _ClassVar[int]
    VIDEOIDLIST_FIELD_NUMBER: _ClassVar[int]
    userInfo: userInfo
    videoIdList: str
    def __init__(self, userInfo: _Optional[_Union[userInfo, _Mapping]] = ..., videoIdList: _Optional[str] = ...) -> None: ...
