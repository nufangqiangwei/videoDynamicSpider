# todo 生成GRPC接口命令
# python -m grpc_tools.protoc -I E:\PythonCode\bilibili_grpc_server --python_out=. --pyi_out=. --grpc_python_out=. E:\PythonCode\bilibili_grpc_server\server.proto
import asyncio
import json
import time
from datetime import datetime

import grpc
from concurrent import futures

from bilibili_api import Credential, video, hot

from bilibili_api.favorite_list import get_video_favorite_list, get_favorite_collected
from bilibili_api.user import User

import server_pb2_grpc

from bili import get_self_user_dynamic, get_self_user_view_history
from collection import get_user_collection, get_user_follow_collection, get_wait_watch_list

from server_pb2 import AuthorInfoResponse, videoInfoResponse, classifyInfoResponse, collectionInfoRequest, \
    collectionInfo

from google.protobuf.json_format import MessageToDict
from google.protobuf import message

from videoDetail import get_video_detail


def grpc_object_to_json(obj):
    if isinstance(obj, message.Message):
        return MessageToDict(obj)


class BilibiliServiceServicer(server_pb2_grpc.WebSiteServiceServicer):
    async def GetUserFollowUpdate(self, request, context):
        sessdata = request.cookies.get("sessdata")
        bili_jct = request.cookies.get('bili_jct')
        buvid3 = request.cookies.get('buvid3')
        dedeuserid = request.cookies.get('dedeuserid')
        ac_time_value = request.cookies.get('ac_time_value')
        request_user_name = request.cookies.get('requestUserName', '')

        client_ip = context.peer()
        start_time = time.time()
        if not all([sessdata, bili_jct, buvid3, dedeuserid]):
            yield videoInfoResponse(
                errorCode=500,
                errorMsg="缺少用户信息",
                requestUserName=request_user_name,
                webSiteName="bilibili",
            )
            return

        yield_response = get_self_user_dynamic(
            sessdata=sessdata,
            bili_jct=bili_jct,
            buvid3=buvid3,
            dedeuserid=dedeuserid,
            ac_time_value=ac_time_value,
            last_update_time=int(request.lastUpdateTime)
        )
        try:
            async for item in yield_response:
                yield item
        except Exception as e:
            print(e)
            yield videoInfoResponse(
                errorCode=500,
                errorMsg="获取动态失败",
            )
            return

        yield videoInfoResponse(
            errorCode=200,
            errorMsg='获取动态完毕',
            requestUserName=request_user_name,
            webSiteName="bilibili",
        )
        end_time = time.time()

    async def GetUserViewHistory(self, request, context):
        client_ip = context.peer()
        start_time = time.time()
        sessdata = request.cookies.get("sessdata")
        bili_jct = request.cookies.get('bili_jct')
        buvid3 = request.cookies.get('buvid3')
        dedeuserid = request.cookies.get('dedeuserid')
        ac_time_value = request.cookies.get('ac_time_value')
        request_user_name = request.cookies.get('requestUserName', '')

        if not all([sessdata, bili_jct, buvid3, dedeuserid]):
            yield videoInfoResponse(
                errorCode=500,
                errorMsg="缺少用户信息",
                requestUserName=request_user_name,
                webSiteName="bilibili",
            )
            return

        yield_response = get_self_user_view_history(
            sessdata,
            bili_jct,
            buvid3,
            dedeuserid,
            ac_time_value,
            last_update_time=int(request.lastHistoryTime),
        )
        async for item in yield_response:
            yield item
        yield videoInfoResponse(
            errorCode=200,
            errorMsg="获取历史记录完毕",
            requestUserName=request_user_name,
            webSiteName="bilibili",
        )
        end_time = time.time()
        print(
            f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 获取历史记录完毕，使用{request_user_name}用户,时间参数是{request.lastHistoryTime}，耗时{int(end_time - start_time)}。")

    async def GetSelfInfo(self, request, context):
        client_ip = context.peer()
        credential = Credential(
            sessdata=request.cookies.get("sessdata"),
            bili_jct=request.cookies.get('bili_jct'),
            buvid3=request.cookies.get('buvid3'),
            dedeuserid=request.cookies.get('dedeuserid'),
            ac_time_value=request.cookies.get('ac_time_value'),
        )
        user = User(request.cookies.get('dedeuserid'), credential=credential)
        user_info = await user.get_user_info()
        print(
            f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 获取用户信息完毕，使用{request.cookies.get('requestUserName')}用户")
        return AuthorInfoResponse(
            name=user_info.get('name', ''),
            avatar=user_info.get('face', ''),
            uid=str(user_info.get('mid', 0)),
            desc=user_info.get('sign', ''),
            followNumber=user_info.get('following', 0),
        )

    async def GetVideoList(self, request, context):
        client_ip = context.peer()
        credential = None
        if request.userInfo is not None:
            credential = Credential(
                sessdata=request.userInfo.cookies.get("sessdata"),
                bili_jct=request.userInfo.cookies.get('bili_jct'),
                buvid3=request.userInfo.cookies.get('buvid3'),
                dedeuserid=request.userInfo.cookies.get('dedeuserid'),
                ac_time_value=request.userInfo.cookies.get('ac_time_value'),
            )

        for wait_video in request.videoIdList:
            yield await self.get_video_list(credential, wait_video)
        print(
            f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 按列表获取视频信息完毕，使用{request.cookies.get('requestUserName')}用户")

    async def GetHotVideoList(self, request, context):
        await hot.get_hot_videos()

    @classmethod
    async def get_video_list(cls, credential, bvid):
        video_obj = video.Video(credential=credential, bvid=bvid)
        video_info = await video_obj.get_detail()
        view_info = video_info.get('View')
        video_info_response = videoInfoResponse(
            title=view_info.get('title'),
            desc=view_info.get('desc'),
            cover=view_info.get('pic'),
            uid=view_info.get('bvid'),
            duration=video_info.get('duration'),
            updateTime=video_info.get('pubdate'),
            viewNumber=view_info.get('stat').get('view'),
            danmaku=view_info.get('stat').get('danmaku'),  # danmaku
            reply=view_info.get('stat').get('reply'),  # reply
            favorite=view_info.get('stat').get('favorite'),  # favorite
            coin=view_info.get('stat').get('coin'),  # coin
            share=view_info.get('stat').get('share'),  # share
            nowRank=view_info.get('stat').get('now_rank'),  # nowRank
            hisRank=view_info.get('stat').get('his_rank'),  # hisRank
            like=view_info.get('stat').get('like'),  # like
            dislike=view_info.get('stat').get('dislike'),  # dislike
            evaluation=view_info.get('stat').get('evaluation'),  # evaluation
        )
        video_info_response.classify.append(classifyInfoResponse(
            name=view_info.get('tname'),
            id=view_info.get('tid'),
        ))
        for tag in video_info.get('Tags'):
            tag_info = video_info_response.tags.add()
            tag_info.name = tag.get('tag_name')
            tag_info.id = tag.get('tag_id')
        if view_info.get('staff'):
            for staff in view_info.get('staff'):
                author_info_response = video_info_response.authors.add()
                author_info_response.name = staff.get('name')
                author_info_response.avatar = staff.get('face')
                author_info_response.uid = str(staff.get('mid'))
                author_info_response.followNumber = staff.get('follower')
        else:
            author_info_response = video_info_response.authors.add()
            author_info_response.name = view_info.get('owner').get('name')
            author_info_response.avatar = view_info.get('owner').get('face')
            author_info_response.uid = str(view_info.get('owner').get('mid'))
            author_info_response.followNumber = video_info.get('Card').get('card').get('fans')
            author_info_response.desc = video_info.get('Card').get('card').get('sign')
        return video_info_response

    async def GetUserFollowList(self, request, context):
        sessdata = request.cookies.get("sessdata")
        bili_jct = request.cookies.get('bili_jct')
        buvid3 = request.cookies.get('buvid3')
        dedeuserid = request.cookies.get('dedeuserid')
        ac_time_value = request.cookies.get('ac_time_value')
        request_user_name = request.cookies.get('requestUserName', '')

        client_ip = context.peer()
        start_time = time.time()
        if not all([sessdata, bili_jct, buvid3, dedeuserid]):
            yield AuthorInfoResponse(
                errorCode=500,
                errorMsg="缺少用户信息",
            )
            return
        print(
            f'{datetime.now().strftime("%Y-%m-%d %H:%M:%S")} {client_ip} 获取用户关注列表完毕，使用{request_user_name}用户')
        credential = Credential(
            sessdata=sessdata,
            bili_jct=bili_jct,
            buvid3=buvid3,
            dedeuserid=dedeuserid,
            ac_time_value=ac_time_value,
        )
        user = User(dedeuserid, credential=credential)
        page = 1
        size = 20
        while True:
            follow_info = await user.get_followings(page, size, False)
            total = follow_info.get('total')
            for author in follow_info.get('list'):
                print(f"{author.get('uname')}")
                yield AuthorInfoResponse(
                    name=author.get('uname'),
                    avatar=author.get('face'),
                    uid=str(author.get('mid')),
                    desc=author.get('sign'),
                    followTime=author.get('mtime'),
                )
            if page * size >= total:
                break
            page += 1
            await asyncio.sleep(0.5)

        end_time = time.time()
        now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        print(f"{now} {client_ip} 获取{request_user_name}用户关注列表耗时 {end_time - start_time} 秒")
        yield AuthorInfoResponse(
            errorCode=200,
            errorMsg="success",
        )

    async def GetUserCollectionList(self, request, context):
        sessdata = request.user.cookies.get("sessdata")
        bili_jct = request.user.cookies.get('bili_jct')
        buvid3 = request.user.cookies.get('buvid3')
        dedeuserid = request.user.cookies.get('dedeuserid')
        request_user_name = request.user.cookies.get('requestUserName', '')

        client_ip = context.peer()
        start_time = time.time()
        error_response = collectionInfo(
            errorCode=500,
            errorMsg="缺少用户信息",
            requestUserName=request_user_name,
            webSiteName="bilibili",
        )

        if not all([sessdata, bili_jct, buvid3, dedeuserid]):
            yield error_response
            return
        credential = Credential(
            sessdata=sessdata,
            bili_jct=bili_jct,
            buvid3=buvid3,
            dedeuserid=dedeuserid,
        )

        # 获取用户创建的收藏夹列表
        try:
            remote_collection_list = await get_video_favorite_list(dedeuserid, credential=credential)
        except Exception as e:
            error_response.errorMsg = str(e)
            yield error_response
            return
        await_get_folder_list = []
        for remote_collect in remote_collection_list.get('list', []):
            await_get_folder_list.append(remote_collect.get('id'))

        # 获取关注的合集列表信息
        page = 1
        result = []
        folder_request = 0
        while True:
            try:
                data = await get_favorite_collected(dedeuserid, pn=page, credential=credential)
            except Exception as e:
                error_response.errorMsg = str(e)
                yield error_response
                return
            page += 1
            result.extend(data.get('list', []))
            for remote_season in data.get('list', []):
                if folder_request < len(await_get_folder_list):
                    yield await get_user_collection(await_get_folder_list[folder_request], credential)
                    await asyncio.sleep(1)
                    folder_request += 1
                yield await get_user_follow_collection(remote_season.get('id'), credential)
            if not data.get('has_more'):
                break
            await asyncio.sleep(0.5)
        while folder_request < len(await_get_folder_list):
            yield await get_user_collection(await_get_folder_list[folder_request], credential)
            folder_request += 1
            await asyncio.sleep(2)
        # 获取稍后观看
        yield await get_wait_watch_list(credential)

        # 执行完成，返回结束信号
        error_response.errorCode = 200
        error_response.errorMsg = "success"
        end_time = time.time()
        now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        print(f"{now} {client_ip} 获取{request_user_name}用户收藏列表耗时 {end_time - start_time} 秒")
        yield error_response

    async def GetVideoDetail(self, request_iterator, context):
        client_ip = context.peer()
        start_time = time.time()
        print(type(request_iterator))
        try:
            async for video_request in request_iterator:
                print(video_request)
                yield await get_video_detail(credential=None, bvid=video_request.videoIdList,aid=None)
        except Exception as e:
            # 处理客户端流结束的异常
            print("客户端流结束的异常")
            print(e)
            pass
        end_time = time.time()
        now = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        print(f"{now} {client_ip} 获取视频信息耗时 {end_time - start_time} 秒")

def create_server():
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    server_pb2_grpc.add_WebSiteServiceServicer_to_server(BilibiliServiceServicer(), server)
    server.add_insecure_port("[::]:50050")
    return server


async def run_server():
    bilibili_server = create_server()
    await bilibili_server.start()
    print("Server started")
    # await bilibili_server.wait_for_termination()
    try:
        # 保持服务器运行
        await asyncio.Future()
    except KeyboardInterrupt:
        # 如果收到键盘中断信号，停止服务器
        await bilibili_server.stop(0)
    print("服务结束")


if __name__ == '__main__':
    # RegisterWebSite('bilibili', '192.168.0.20', 50051)
    try:
        asyncio.run(run_server(), debug=True)
    finally:
        # InvalidGrpcClient('bilibili', '192.168.0.20', 50051)
        pass
