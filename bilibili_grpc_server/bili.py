from bilibili_api import Credential, user, dynamic
from bilibili_api.user import get_self_history_new, HistoryType, HistoryBusinessType
from bilibili_api.utils.network import Api

from server_pb2 import videoInfoResponse, viewInfoResponse


async def download_self_dynamic(credential: Credential, offset=None):
    api = dynamic.API["info"]["dynamic_page_info"]
    params = {
        "timezone_offset": -480,
        "features": "itemOpusStyle",
        "page": 1,
        "type": "video"
    }
    params.update({"offset": offset} if offset else {})
    dynamic_data = (
        await Api(**api, credential=credential).update_params(**params).result
    )
    return dynamic_data


async def get_self_user_dynamic(sessdata, bili_jct, buvid3, dedeuserid, ac_time_value=None, last_update_time=None):
    credential = Credential(sessdata=sessdata,
                            bili_jct=bili_jct,
                            buvid3=buvid3,
                            dedeuserid=dedeuserid,
                            ac_time_value=ac_time_value,
                            )
    offset = None
    video_number = 20
    while True:
        follow_video_list = await download_self_dynamic(credential=credential, offset=offset)
        offset = follow_video_list.get('offset')
        for follow_video in follow_video_list.get('items'):
            if follow_video.get('type') != 'DYNAMIC_TYPE_AV':
                continue
            if last_update_time is not None and follow_video.get("modules").get("module_author").get('pub_ts') <= last_update_time:
                return
            yield build_dynamic_video_info_response(follow_video)
            if last_update_time is None:
                video_number -= 1
            if video_number == 0:
                return


def build_dynamic_video_info_response(data: dict) -> videoInfoResponse:
    video_info = data.get("modules").get("module_dynamic").get("major").get("archive")
    author_info = data.get("modules").get("module_author")
    video_info_response = videoInfoResponse(
        title=video_info.get("title"),
        desc=video_info.get("desc"),
        cover=video_info.get("cover"),
        uid=video_info.get("bvid"),
        duration=HourAndMinutesAndSecondsToSeconds(video_info.get("duration_text")),
        updateTime=author_info.get("pub_ts"),
        webSiteName="bilibili"
    )
    author_info_response = video_info_response.authors.add()
    author_info_response.name = author_info.get("name")
    author_info_response.avatar = author_info.get("face")
    author_info_response.uid = str(author_info.get("mid"))
    
    return video_info_response


async def get_self_user_view_history(sessdata, bili_jct, buvid3, dedeuserid, ac_time_value=None, last_update_time=None):
    credential = Credential(sessdata=sessdata,
                            bili_jct=bili_jct,
                            buvid3=buvid3,
                            dedeuserid=dedeuserid,
                            ac_time_value=ac_time_value,
                            )
    view_at = 0
    video_number = 10
    while True:
        view_history_data = await get_self_history_new(credential=credential, _type=HistoryType.archive,
                                                       business=HistoryBusinessType.archive, view_at=view_at)
        
        for view_history in view_history_data.get('list'):
            if view_history.get('badge') != '':
                continue
            if last_update_time is not None and view_history.get("view_at") <= last_update_time:
                return
            if last_update_time is None:
                video_number -= 1
            if video_number == 0:
                return
            view_at = view_history.get("view_at")
            yield build_history_video_info_response(view_history)


def build_history_video_info_response(data: dict) -> videoInfoResponse:
    video_info_response = videoInfoResponse(
        title=data.get("title"),
        desc=data.get("history").get("part"),
        cover=data.get("cover"),
        uid=data.get("history").get("bvid"),
        duration=data.get("duration"),
        viewInfo=viewInfoResponse(
            viewTime=data.get("view_at"),
            viewDuration=data.get("progress"),
        ),
    )
    author_info_response = video_info_response.authors.add()
    author_info_response.name = data.get("author_name")
    author_info_response.avatar = data.get("author_face")
    author_info_response.uid = str(data.get("author_mid"))
    
    # video_info_response.viewInfo = viewInfoResponse(
    #     viewTime=data.get("view_at"),
    #     viewDuration=data.get("progress"),
    # )
    
    return video_info_response


def HourAndMinutesAndSecondsToSeconds(time_str: str) -> int:
    split = time_str.split(":")
    if len(split) == 2:
        minutes = int(split[0])
        seconds = int(split[1])
        return minutes * 60 + seconds
    elif len(split) == 3:
        hour = int(split[0])
        minutes = int(split[1])
        seconds = int(split[2])
        return hour * 3600 + minutes * 60 + seconds
    return 0


if __name__ == '__main__':
    dedeuserid = "aw"

