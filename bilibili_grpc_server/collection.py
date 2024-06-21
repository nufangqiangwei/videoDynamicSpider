import asyncio
import random
from typing import List

from bilibili_api import Credential
from bilibili_api.exceptions import ResponseCodeException
from bilibili_api.favorite_list import get_video_favorite_list_content
from bilibili_api.user import get_toview_list, User
from bilibili_api.channel_series import ChannelSeries, API_USER
from bilibili_api.favorite_list import get_video_favorite_list, get_favorite_collected
from server_pb2 import collectionInfo, videoInfoResponse, AuthorInfoResponse
from bilibili_api.utils.network import Api


# 个人收藏夹
async def get_user_collection(collection_id, credential):
    # 获取精简版的视频列表，只包含视频id,不过一次性就能获取完整的列表
    # collection = FavoriteList(media_id=collection_id, credential=credential)
    # await collection.get_content_ids_info()
    page = 1
    try:
        response = await get_video_favorite_list_content(collection_id, page=page, credential=credential)
    except Exception as e:
        print(e)
        return make_error_videoInfoResponse('folder', collection_id, str(e))
    new_collection = make_videoInfoResponse(response, 'folder')
    while response.get('has_more'):
        try:
            response = await get_video_favorite_list_content(collection_id, page=page, credential=credential)
        except Exception as e:
            print(e)
            new_collection.errorCode = 500
            new_collection.errorMsg = str(e)
            return new_collection
        append_videoInfoResponse_video(new_collection, response)
        # 获取个2.5到4的数字 休眠
        await asyncio.sleep(random.uniform(2.5, 4))
        page += 1
    return new_collection

# 用户稍后观看列表
async def get_wait_watch_list(credential:Credential):
    response = await get_toview_list(credential)
    new_collection = collectionInfo(
        collectionId=int(credential.dedeuserid),
        name='稍后观看',
        collectionType='waitWatch',
        upperUid=credential.dedeuserid,
    )
    for i in response.get('list', []):
        video_info = videoInfoResponse(
            title=i.get('title'),
            cover=i.get('pic'),
            uid=i.get('bvid'),
            duration=i.get('duration'),
            updateTime=i.get('ctime'),
            collectTime=i.get('add_at'),
            viewNumber=i.get('stat', {}).get('view'),
            danmaku=i.get('stat', {}).get('danmaku'),
            reply=i.get('stat', {}).get('reply'),
        )
        video_info.authors.append(AuthorInfoResponse(
            name=i.get('owner', {}).get('name'),
            avatar=i.get('owner', {}).get('face'),
            uid=str(i.get('owner', {}).get('mid')),
        ))
        new_collection.video.append(video_info)
    return new_collection

# 用户关注的合集
async def get_user_follow_collection(collection_id, credential):
    api = API_USER["channel_series"]["season_info"]
    params = {"season_id": collection_id}
    # 一次性返回整个的列表
    try:
        response = await Api(**api, credential=credential).update_params(**params).result
    except ResponseCodeException as e:
        error = make_error_videoInfoResponse('subscription', collection_id, str(e))
        error.errorCode = e.code
        return error
    except Exception as e:
        print(e)
        return make_error_videoInfoResponse('subscription', collection_id, str(e))
    await asyncio.sleep(0.5)
    return make_videoInfoResponse(response, "subscription")


def make_videoInfoResponse(response, collectionType):
    new_collection = collectionInfo(
        collectionId=response.get('info', {}).get('id'),
        name=response.get('info', {}).get('title'),
        collectionType=collectionType,
        upperUid=str(response.get('info', {}).get('upper', {}).get('mid')),
        upperName=response.get('info', {}).get('upper', {}).get('name'),
    )
    append_videoInfoResponse_video(new_collection, response)
    return new_collection


def append_videoInfoResponse_video(new_collection, response):
    for i in response.get('medias', []):
        video_info = videoInfoResponse(
            title=i.get('title'),
            cover=i.get('cover'),
            uid=i.get('bvid'),
            duration=i.get('duration'),
            updateTime=i.get('ctime'),
            collectTime=i.get('fav_time'),
            viewNumber=i.get('cnt_info', {}).get('play'),
            danmaku=i.get('cnt_info', {}).get('danmaku'),
            reply=i.get('cnt_info', {}).get('reply'),
            favorite=i.get('cnt_info', {}).get('collect'),
            IsInvalid=i.get('attr', 0) == 9,
        )
        video_info.authors.append(AuthorInfoResponse(
            name=i.get('upper', {}).get('name'),
            avatar=i.get('upper', {}).get('face'),
            uid=str(i.get('upper', {}).get('mid')),
        ))
        new_collection.video.append(video_info)


def make_error_videoInfoResponse(collectionType, collection_id, errorMsg):
    new_collection = collectionInfo(
        collectionId=collection_id,
        collectionType=collectionType,
        errorCode=500,
        errorMsg=errorMsg,
    )
    return new_collection



