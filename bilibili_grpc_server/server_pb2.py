# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: server.proto
# Protobuf Python Version: 4.25.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0cserver.proto\x12\x0bwebSiteGRPC\"\xa0\x01\n\x08userInfo\x12\x33\n\x07\x63ookies\x18\x01 \x03(\x0b\x32\".webSiteGRPC.userInfo.CookiesEntry\x12\x17\n\x0flastHistoryTime\x18\x02 \x01(\t\x12\x16\n\x0elastUpdateTime\x18\x03 \x01(\t\x1a.\n\x0c\x43ookiesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\x93\x05\n\x11videoInfoResponse\x12\r\n\x05title\x18\x01 \x01(\t\x12\x0c\n\x04\x64\x65sc\x18\x02 \x01(\t\x12\r\n\x05\x63over\x18\x03 \x01(\t\x12\x0b\n\x03uid\x18\x04 \x01(\t\x12\x10\n\x08\x64uration\x18\x05 \x01(\x03\x12\x12\n\nupdateTime\x18\x06 \x01(\x03\x12\x13\n\x0b\x63ollectTime\x18\x17 \x01(\x03\x12*\n\x04tags\x18\x07 \x03(\x0b\x32\x1c.webSiteGRPC.tagInfoResponse\x12\x33\n\x08\x63lassify\x18\x08 \x01(\x0b\x32!.webSiteGRPC.classifyInfoResponse\x12\x12\n\nviewNumber\x18\t \x01(\x03\x12\x0f\n\x07\x64\x61nmaku\x18\n \x01(\x03\x12\r\n\x05reply\x18\x0b \x01(\x03\x12\x10\n\x08\x66\x61vorite\x18\x0c \x01(\x03\x12\x0c\n\x04\x63oin\x18\r \x01(\x03\x12\r\n\x05share\x18\x0e \x01(\x03\x12\x0f\n\x07nowRank\x18\x0f \x01(\x03\x12\x0f\n\x07hisRank\x18\x10 \x01(\x03\x12\x0c\n\x04like\x18\x11 \x01(\x03\x12\x0f\n\x07\x64islike\x18\x12 \x01(\x03\x12\x12\n\nevaluation\x18\x13 \x01(\t\x12\x30\n\x07\x61uthors\x18\x14 \x03(\x0b\x32\x1f.webSiteGRPC.AuthorInfoResponse\x12/\n\x08viewInfo\x18\x15 \x01(\x0b\x32\x1d.webSiteGRPC.viewInfoResponse\x12\x13\n\x0bwebSiteName\x18\x16 \x01(\t\x12\x11\n\tIsInvalid\x18\x18 \x01(\x08\x12\x11\n\terrorCode\x18\x32 \x01(\x05\x12\x10\n\x08\x65rrorMsg\x18\x33 \x01(\t\x12\x17\n\x0frequestUserName\x18\x34 \x01(\t\x12\x15\n\rrequestUserId\x18\x35 \x01(\x03\x12\x11\n\twebSiteId\x18\x36 \x01(\x03\"\x84\x02\n\x12\x41uthorInfoResponse\x12\x0e\n\x06\x61uthor\x18\x01 \x01(\t\x12\x0c\n\x04name\x18\x02 \x01(\t\x12\x0e\n\x06\x61vatar\x18\x03 \x01(\t\x12\x0c\n\x04\x64\x65sc\x18\x04 \x01(\t\x12\x0b\n\x03uid\x18\x05 \x01(\t\x12\x14\n\x0c\x66ollowNumber\x18\x06 \x01(\x04\x12\x12\n\nfollowTime\x18\x07 \x01(\x03\x12\x13\n\x0bwebSiteName\x18\x16 \x01(\t\x12\x11\n\terrorCode\x18\x32 \x01(\x05\x12\x10\n\x08\x65rrorMsg\x18\x33 \x01(\t\x12\x17\n\x0frequestUserName\x18\x34 \x01(\t\x12\x15\n\rrequestUserId\x18\x35 \x01(\x03\x12\x11\n\twebSiteId\x18\x36 \x01(\x03\"a\n\x0ftagInfoResponse\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\n\n\x02id\x18\x02 \x01(\x03\x12\x0f\n\x07tagType\x18\x03 \x01(\x03\x12\x11\n\terrorCode\x18\x32 \x01(\x05\x12\x10\n\x08\x65rrorMsg\x18\x33 \x01(\t\"0\n\x14\x63lassifyInfoResponse\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\n\n\x02id\x18\x02 \x01(\x03\":\n\x10viewInfoResponse\x12\x10\n\x08viewTime\x18\x01 \x01(\x03\x12\x14\n\x0cviewDuration\x18\x02 \x01(\x03\"\xb0\x01\n\x16\x63ollectionInfoResponse\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0b\n\x03uid\x18\x02 \x01(\t\x12/\n\x06\x61uthor\x18\x03 \x01(\x0b\x32\x1f.webSiteGRPC.AuthorInfoResponse\x12\x11\n\tcreatTime\x18\x04 \x01(\x03\x12\x12\n\nappendTime\x18\x05 \x01(\x03\x12\x11\n\terrorCode\x18\x32 \x01(\x05\x12\x10\n\x08\x65rrorMsg\x18\x33 \x01(\t\"\xe8\x02\n\x0e\x63ollectionInfo\x12\x16\n\x0e\x63ollectionType\x18\x01 \x01(\t\x12\x14\n\x0c\x63ollectionId\x18\x02 \x01(\x03\x12\x1c\n\x14\x63ollectionVideoCount\x18\x03 \x01(\x03\x12\x16\n\x0elastAppendTime\x18\x04 \x01(\x03\x12\x0c\n\x04name\x18\x05 \x01(\t\x12\x11\n\tupperName\x18\x06 \x01(\t\x12\x10\n\x08upperUid\x18\x07 \x01(\t\x12-\n\x05video\x18\x14 \x03(\x0b\x32\x1e.webSiteGRPC.videoInfoResponse\x12\x13\n\x0bvideoNumber\x18\x15 \x01(\x03\x12\x13\n\x0bwebSiteName\x18\x31 \x01(\t\x12\x11\n\terrorCode\x18\x32 \x01(\x05\x12\x10\n\x08\x65rrorMsg\x18\x33 \x01(\t\x12\x17\n\x0frequestUserName\x18\x34 \x01(\t\x12\x15\n\rrequestUserId\x18\x35 \x01(\x03\x12\x11\n\twebSiteId\x18\x36 \x01(\x03\"m\n\x15\x63ollectionInfoRequest\x12#\n\x04user\x18\x01 \x01(\x0b\x32\x15.webSiteGRPC.userInfo\x12/\n\ncollection\x18\x02 \x03(\x0b\x32\x1b.webSiteGRPC.collectionInfo\"S\n\x13getVideoListRequest\x12\'\n\x08userInfo\x18\x01 \x01(\x0b\x32\x15.webSiteGRPC.userInfo\x12\x13\n\x0bvideoIdList\x18\x02 \x01(\t\"\x82\x01\n\x13VideoDetailResponse\x12\x33\n\x0bvideoDetail\x18\x01 \x01(\x0b\x32\x1e.webSiteGRPC.videoInfoResponse\x12\x36\n\x0erecommendVideo\x18\x02 \x03(\x0b\x32\x1e.webSiteGRPC.videoInfoResponse2\x9a\x05\n\x0eWebSiteService\x12N\n\x13GetUserFollowUpdate\x12\x15.webSiteGRPC.userInfo\x1a\x1e.webSiteGRPC.videoInfoResponse0\x01\x12M\n\x12GetUserViewHistory\x12\x15.webSiteGRPC.userInfo\x1a\x1e.webSiteGRPC.videoInfoResponse0\x01\x12M\n\x11GetUserFollowList\x12\x15.webSiteGRPC.userInfo\x1a\x1f.webSiteGRPC.AuthorInfoResponse0\x01\x12Z\n\x15GetUserCollectionList\x12\".webSiteGRPC.collectionInfoRequest\x1a\x1b.webSiteGRPC.collectionInfo0\x01\x12U\n\x0fGetHotVideoList\x12 .webSiteGRPC.getVideoListRequest\x1a\x1e.webSiteGRPC.videoInfoResponse0\x01\x12\x45\n\x0bGetSelfInfo\x12\x15.webSiteGRPC.userInfo\x1a\x1f.webSiteGRPC.AuthorInfoResponse\x12\x46\n\x10GetWaitWatchList\x12\x15.webSiteGRPC.userInfo\x1a\x1b.webSiteGRPC.collectionInfo\x12X\n\x0eGetVideoDetail\x12 .webSiteGRPC.getVideoListRequest\x1a .webSiteGRPC.VideoDetailResponse(\x01\x30\x01\x42%Z#videoDynamicAcquisition/proto;protob\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'server_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  _globals['DESCRIPTOR']._options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z#videoDynamicAcquisition/proto;proto'
  _globals['_USERINFO_COOKIESENTRY']._options = None
  _globals['_USERINFO_COOKIESENTRY']._serialized_options = b'8\001'
  _globals['_USERINFO']._serialized_start=30
  _globals['_USERINFO']._serialized_end=190
  _globals['_USERINFO_COOKIESENTRY']._serialized_start=144
  _globals['_USERINFO_COOKIESENTRY']._serialized_end=190
  _globals['_VIDEOINFORESPONSE']._serialized_start=193
  _globals['_VIDEOINFORESPONSE']._serialized_end=852
  _globals['_AUTHORINFORESPONSE']._serialized_start=855
  _globals['_AUTHORINFORESPONSE']._serialized_end=1115
  _globals['_TAGINFORESPONSE']._serialized_start=1117
  _globals['_TAGINFORESPONSE']._serialized_end=1214
  _globals['_CLASSIFYINFORESPONSE']._serialized_start=1216
  _globals['_CLASSIFYINFORESPONSE']._serialized_end=1264
  _globals['_VIEWINFORESPONSE']._serialized_start=1266
  _globals['_VIEWINFORESPONSE']._serialized_end=1324
  _globals['_COLLECTIONINFORESPONSE']._serialized_start=1327
  _globals['_COLLECTIONINFORESPONSE']._serialized_end=1503
  _globals['_COLLECTIONINFO']._serialized_start=1506
  _globals['_COLLECTIONINFO']._serialized_end=1866
  _globals['_COLLECTIONINFOREQUEST']._serialized_start=1868
  _globals['_COLLECTIONINFOREQUEST']._serialized_end=1977
  _globals['_GETVIDEOLISTREQUEST']._serialized_start=1979
  _globals['_GETVIDEOLISTREQUEST']._serialized_end=2062
  _globals['_VIDEODETAILRESPONSE']._serialized_start=2065
  _globals['_VIDEODETAILRESPONSE']._serialized_end=2195
  _globals['_WEBSITESERVICE']._serialized_start=2198
  _globals['_WEBSITESERVICE']._serialized_end=2864
# @@protoc_insertion_point(module_scope)
