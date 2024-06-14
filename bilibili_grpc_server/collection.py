from typing import List

from bilibili_api import Credential
from bilibili_api.favorite_list import get_video_favorite_list_content
from bilibili_api.user import get_toview_list
from bilibili_api.channel_series import ChannelSeries, API_USER
from bilibili_api.favorite_list import get_video_favorite_list, get_favorite_collected
from server_pb2 import collectionInfo, videoInfoResponse, AuthorInfoResponse
from bilibili_api.utils.network import Api


async def update_user_collection_info(credential: Credential, db_collection_list: List[collectionInfo]):
    # 获取用户创建的收藏夹列表
    remote_collection_list = await get_video_favorite_list(credential.dedeuserid, credential=credential)
    for remote_collect in remote_collection_list.get('list', []):
        local_collect = None
        for db_collection in db_collection_list:
            if db_collection.collectionId == str(remote_collect.get('id')):
                local_collect = db_collection
                break
        yield await get_user_collection(remote_collect.get('id'), credential, local_collect is None)

    # 获取关注的合集列表信息
    page = 1
    result = []
    while True:
        data = await get_favorite_collected(credential.dedeuserid, pn=page, credential=credential)
        page += 1
        result.extend(data.get('list', []))
        for remote_season in data.get('list', []):
            yield await get_user_follow_collection(remote_season.get('id'), credential)
        if not data.get('has_more'):
            break


# 个人收藏夹
async def get_user_collection(collection_id, credential, get_all=False):
    # 获取精简版的视频列表，只包含视频id,不过一次性就能获取完整的列表
    # collection = FavoriteList(media_id=collection_id, credential=credential)
    # await collection.get_content_ids_info()
    page = 1
    response = await get_video_favorite_list_content(collection_id, page=page, credential=credential)
    new_collection = collectionInfo(
        collectionId=str(response.get('info', {}).get('id')),
        name=response.get('info', {}).get('title'),
        collectionType="folder",
        upperUid=str(response.get('info', {}).get('upper', {}).get('mid')),
        upperName=response.get('info', {}).get('upper', {}).get('name'),
    )
    while response.get('has_more') or page == 1:
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
            )
            video_info.authors.append(AuthorInfoResponse(
                name=i.get('upper', {}).get('name'),
                avatar=i.get('upper', {}).get('face'),
                uid=str(i.get('upper', {}).get('mid')),
            ))
            new_collection.video.append(video_info)
        if not get_all:
            break
        response = await get_video_favorite_list_content(collection_id, page=page, credential=credential)
        page += 1
    return new_collection


# 个人收藏夹
async def get_user_wait_watch(credential):
    await get_toview_list(credential)


# 用户关注的合集
async def get_user_follow_collection(collection_id, credential):
    api = API_USER["channel_series"]["season_info"]
    params = {"season_id": collection_id}
    # 一次性返回整个的列表
    return (
        await Api(**api, credential=credential).update_params(**params).result
    )
