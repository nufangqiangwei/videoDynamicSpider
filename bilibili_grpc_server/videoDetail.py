from bilibili_api import sync, Credential, homepage, hot, video
from bilibili_api.utils.network import Api
from bilibili_api.utils.utils import get_api

from server_pb2 import VideoDetailResponse, videoInfoResponse, AuthorInfoResponse, tagInfoResponse

video_API = get_api("video")


async def get_video_detail(credential: Credential, bvid: str, aid: int) -> VideoDetailResponse:
    api = video_API["info"]["detail"]
    params = {"bvid": bvid, "aid": aid}
    response = await Api(**api, credential=credential).update_params(**params).request(proxy="http://127.0.0.1:1080")
    remote_video_info = response.get("View", {})
    video_detail_response = VideoDetailResponse(
        videoDetail=gen_video_detail_response(remote_video_info)
    )

    for related_info in response.get('Related', []):
        video_detail_response.recommendVideo.append(gen_video_detail_response(related_info))
    return video_detail_response


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


def gen_video_detail_response(response) -> videoInfoResponse:
    view_info = response.get('stat', {})
    video_detail = videoInfoResponse(
        title=response.get("title", ''),
        desc=response.get("desc", ''),
        cover=response.get("pic", ''),
        uid=response.get("bvid", ''),
        duration=response.get("duration", 0),
        updateTime=response.get("pubdate", 0),
        viewNumber=view_info.get('view', 0),
        danmaku=view_info.get("danmaku", 0),
        reply=view_info.get("reply", 0),
        favorite=view_info.get("favorite", 0),
        coin=view_info.get("coin", 0),
        share=view_info.get("share", 0),
        nowRank=view_info.get("now_rank", 0),
        hisRank=view_info.get("his_rank", 0),
        like=view_info.get("like", 0),
        dislike=view_info.get("dislike", 0),
        evaluation=view_info.get("evaluation", '0'),
    )
    if response.get('staff'):
        for staff in response.get('staff'):
            video_detail.authors.append(AuthorInfoResponse(
                name=staff.get('name', ''),
                uid=str(staff.get('mid', '')),
                avatar=staff.get('face', ''),
                author=staff.get('title', ''),
            ))
    elif response.get('owner'):
        video_detail.authors.append(AuthorInfoResponse(
            name=response.get("owner", {}).get("name", ''),
            uid=str(response.get("owner", {}).get("mid", '')),
            avatar=response.get("owner", {}).get("face", ''),
        ))

    for tag in response.get("Tags", []):
        if tag.get('tag_name') in response.get("participle", []):
            video_detail.tags.append(tagInfoResponse(
                name=tag.get("tag_name"),
                id=tag.get("tag_id"),
                tagType=1,
            ))
        else:
            video_detail.tags.append(tagInfoResponse(
                name=tag.get("tag_name"),
                id=tag.get("tag_id"),
                tagType=3,
            ))

    return video_detail
