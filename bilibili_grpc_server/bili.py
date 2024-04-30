from bilibili_api import Credential, user, dynamic
from bilibili_api.user import get_self_history_new, HistoryType, HistoryBusinessType
from bilibili_api.utils.network import Api

from server_pb2 import videoInfoResponse, viewInfoResponse

dynamic_json = '''
{
    "basic": {
        "comment_id_str": "1803789226",
        "comment_type": 1,
        "like_icon": {
            "action_url": "",
            "end_url": "",
            "id": 0,
            "start_url": ""
        },
        "rid_str": "1803789226"
    },
    "id_str": "925271161134120979",
    "modules": {
        "module_author": {
            "avatar": {
                "container_size": {
                    "height": 1.35,
                    "width": 1.35
                },
                "fallback_layers": {
                    "is_critical_group": true,
                    "layers": [
                        {
                            "general_spec": {
                                "pos_spec": {
                                    "axis_x": 0.675,
                                    "axis_y": 0.675,
                                    "coordinate_pos": 2
                                },
                                "render_spec": {
                                    "opacity": 1
                                },
                                "size_spec": {
                                    "height": 1,
                                    "width": 1
                                }
                            },
                            "layer_config": {
                                "is_critical": true,
                                "tags": {
                                    "AVATAR_LAYER": {},
                                    "GENERAL_CFG": {
                                        "config_type": 1,
                                        "general_config": {
                                            "web_css_style": {
                                                "borderRadius": "50%"
                                            }
                                        }
                                    }
                                }
                            },
                            "resource": {
                                "res_image": {
                                    "image_src": {
                                        "placeholder": 6,
                                        "remote": {
                                            "bfs_style": "widget-layer-avatar",
                                            "url": "https://i0.hdslb.com/bfs/face/f1f2f7a22549905c69209439602d661c1b2a1975.jpg"
                                        },
                                        "src_type": 1
                                    }
                                },
                                "res_type": 3
                            },
                            "visible": true
                        },
                        {
                            "general_spec": {
                                "pos_spec": {
                                    "axis_x": 0.8000000000000002,
                                    "axis_y": 0.8000000000000002,
                                    "coordinate_pos": 1
                                },
                                "render_spec": {
                                    "opacity": 1
                                },
                                "size_spec": {
                                    "height": 0.41666666666666663,
                                    "width": 0.41666666666666663
                                }
                            },
                            "layer_config": {
                                "tags": {
                                    "GENERAL_CFG": {
                                        "config_type": 1,
                                        "general_config": {
                                            "web_css_style": {
                                                "background-color": "rgb(255,255,255)",
                                                "border": "2px solid rgba(255,255,255,1)",
                                                "borderRadius": "50%",
                                                "boxSizing": "border-box"
                                            }
                                        }
                                    },
                                    "ICON_LAYER": {}
                                }
                            },
                            "resource": {
                                "res_image": {
                                    "image_src": {
                                        "local": 1,
                                        "src_type": 2
                                    }
                                },
                                "res_type": 3
                            },
                            "visible": true
                        }
                    ]
                },
                "mid": "432106248"
            },
            "face": "https://i0.hdslb.com/bfs/face/f1f2f7a22549905c69209439602d661c1b2a1975.jpg",
            "face_nft": false,
            "following": true,
            "jump_url": "//space.bilibili.com/432106248/dynamic",
            "label": "",
            "mid": 432106248,
            "name": "翟萌萌同学",
            "official_verify": {
                "desc": "",
                "type": -1
            },
            "pendant": {
                "expire": 0,
                "image": "",
                "image_enhance": "",
                "image_enhance_frame": "",
                "n_pid": 0,
                "name": "",
                "pid": 0
            },
            "pub_action": "投稿了视频",
            "pub_location_text": "",
            "pub_time": "13分钟前",
            "pub_ts": 1714269880,
            "type": "AUTHOR_TYPE_NORMAL",
            "vip": {
                "avatar_subscript": 1,
                "avatar_subscript_url": "",
                "due_date": 1728144000000,
                "label": {
                    "bg_color": "#FB7299",
                    "bg_style": 1,
                    "border_color": "",
                    "img_label_uri_hans": "",
                    "img_label_uri_hans_static": "https://i0.hdslb.com/bfs/vip/8d4f8bfc713826a5412a0a27eaaac4d6b9ede1d9.png",
                    "img_label_uri_hant": "",
                    "img_label_uri_hant_static": "https://i0.hdslb.com/bfs/activity-plat/static/20220614/e369244d0b14644f5e1a06431e22a4d5/VEW8fCC0hg.png",
                    "label_theme": "annual_vip",
                    "path": "",
                    "text": "年度大会员",
                    "text_color": "#FFFFFF",
                    "use_img_label": true
                },
                "nickname_color": "#FB7299",
                "status": 1,
                "theme_type": 0,
                "type": 2
            }
        },
        "module_dynamic": {
            "additional": null,
            "desc": null,
            "major": {
                "archive": {
                    "aid": "1803789226",
                    "badge": {
                        "bg_color": "#FB7299",
                        "color": "#FFFFFF",
                        "icon_url": null,
                        "text": "投稿视频"
                    },
                    "bvid": "BV1Mb421e7Ui",
                    "cover": "http://i1.hdslb.com/bfs/archive/38c50b1d3b151fca35fbbd7d1bd7dcdd545fcf06.jpg",
                    "desc": "",
                    "disable_preview": 0,
                    "duration_text": "02:23",
                    "jump_url": "//www.bilibili.com/video/BV1Mb421e7Ui/",
                    "stat": {
                        "danmaku": "0",
                        "play": "93"
                    },
                    "title": "据说大家都有个朋友，在用它的名字做昵称。【半夏】",
                    "type": 1
                },
                "type": "MAJOR_TYPE_ARCHIVE"
            },
            "topic": null
        },
        "module_more": {
            "three_point_items": [
                {
                    "label": "取消关注",
                    "type": "THREE_POINT_FOLLOWING"
                },
                {
                    "label": "举报",
                    "type": "THREE_POINT_REPORT"
                }
            ]
        },
        "module_stat": {
            "comment": {
                "count": 3,
                "forbidden": false
            },
            "forward": {
                "count": 0,
                "forbidden": false
            },
            "like": {
                "count": 16,
                "forbidden": false,
                "status": false
            }
        }
    },
    "type": "DYNAMIC_TYPE_AV",
    "visible": true
}
'''


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
        print(follow_video_list)
        offset = follow_video_list.get('offset')
        for follow_video in follow_video_list.get('items'):
            if follow_video.get('type') != 'DYNAMIC_TYPE_AV':
                continue
            if last_update_time is not None and follow_video.get("modules").get("module_author").get(
                    'pub_ts') <= last_update_time:
                return
            if last_update_time is None:
                video_number -= 1
            if video_number == 0:
                return

            yield build_dynamic_video_info_response(follow_video)


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
    video_number = 100
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
        )
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
    print(HourAndMinutesAndSecondsToSeconds("02:23"))
