from bilibili_api import sync, Credential, homepage, hot,video
from bilibili_api.utils.network import Api
from bilibili_api.utils.utils import get_api

from server_pb2 import VideoDetailResponse, videoInfoResponse, AuthorInfoResponse, tagInfoResponse

video_API = get_api("video")
async def get_video_detail(credential: Credential, bvid: str, aid: int):
    api = video_API["info"]["detail"]
    params = {"bvid": bvid, "aid": aid}
    response = await Api(**api, credential=credential).update_params(**params).request(proxy="http://127.0.0.1:7890")
    video_detail_response = VideoDetailResponse()
    remote_video_info = response.get("View", {})
    video_detail = videoInfoResponse(
        title=remote_video_info["title"],
        desc=remote_video_info["desc"],
        cover=remote_video_info["pic"],
        uid=remote_video_info["bvid"],
        duration=remote_video_info["duration"],
        updateTime=remote_video_info["pubdate"],
        viewNumber=remote_video_info["aid"],
        danmaku=remote_video_info["stat"]["danmaku"],
        reply=remote_video_info["stat"]["reply"],
        favorite=remote_video_info["stat"]["favorite"],
        coin=remote_video_info["stat"]["coin"],
        share=remote_video_info["stat"]["share"],
        nowRank=remote_video_info["stat"]["now_rank"],
        hisRank=remote_video_info["stat"]["his_rank"],
        like=remote_video_info["stat"]["like"],
        dislike=remote_video_info["stat"]["dislike"],
        evaluation=remote_video_info["stat"]["evaluation"],
    )

    if response.get('View', {}).get('staff'):
        for staff in response.get('View',{}).get('staff'):
            video_detail.authors.append(AuthorInfoResponse(
                name=staff['name'],
                uid=staff['mid'],
                avatar=staff['face'],
                author=staff['title'],
            ))
    else:
        video_detail.authors.append(AuthorInfoResponse(
            name=remote_video_info["owner"]["name"],
            uid=remote_video_info["owner"]["mid"],
            avatar=remote_video_info["owner"]["face"],
        ))

    for tag in response["Tags"]:
        if tag.get('tag_name') in response["participle"]:
            video_detail.tags.append(tagInfoResponse(
                name=tag["tag_name"],
                id=tag["tag_id"],
                tagType=1,
            ))
        else:
            video_detail.tags.append(tagInfoResponse(
                name=tag["tag_name"],
                id=tag["tag_id"],
                tagType=3,
            ))


# 获取是竖屏还是横屏视频
def video_aspect_ratio(width: int, height: int) -> float:
    """
    获取是竖屏还是横屏视频
    1:1和9:16之间,则为竖屏视频。比例接近16:9或更宽,则为横屏视频。
    :param width:
    :param height:
    :return:
    """
    ratio = width / height
    if ratio >= 0.9 and ratio <= 1.1:
        return 1
    else:
        return 0


