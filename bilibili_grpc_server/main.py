# todo 生成GRPC接口命令
# python -m grpc_tools.protoc -I E:\PythonCode\bilibili_grpc_server --python_out=. --pyi_out=. --grpc_python_out=. E:\PythonCode\bilibili_grpc_server\server.proto
import asyncio
import time
from datetime import datetime

import grpc
from concurrent import futures

from bilibili_api import Credential, video, hot
from bilibili_api.user import User

import server_pb2_grpc

from bili import get_self_user_dynamic, get_self_user_view_history
from redisDiscovery import RegisterWebSite, InvalidGrpcClient
from server_pb2 import AuthorInfoResponse, videoInfoResponse, classifyInfoResponse


class BilibiliServiceServicer(server_pb2_grpc.WebSiteServiceServicer):
    async def GetUserFollowUpdate(self, request, context):
        client_ip = context.peer()
        start_time = time.time()
        yield_response = get_self_user_dynamic(
            request.cookies.get("sessdata"),
            request.cookies.get('bili_jct'),
            request.cookies.get('buvid3'),
            request.cookies.get('dedeuserid'),
            request.cookies.get('ac_time_value'),
            last_update_time=int(request.lastUpdateTime)
        )
        index = 0
        last_get_time = ''
        async for item in yield_response:
            if index == 0:
                last_get_time = item.updateTime
            index += 1
            await asyncio.sleep(0.5)
            yield item
        yield videoInfoResponse(
            errorCode=200,
            errorMsg=str(last_get_time),
            requestUserName=request.cookies.get("requestUserName", ""),
            webSiteName="bilibili",
        )
        end_time = time.time()
        print(f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 获取动态完毕，使用{request.cookies.get('requestUserName')}用户,时间参数是{request.lastUpdateTime}，耗时{int(end_time-start_time)}。获取到{index}个数据")
    
    async def GetUserViewHistory(self, request, context):
        client_ip = context.peer()
        start_time = time.time()
        yield_response = get_self_user_view_history(
            request.cookies.get("sessdata"),
            request.cookies.get('bili_jct'),
            request.cookies.get('buvid3'),
            request.cookies.get('dedeuserid'),
            request.cookies.get('ac_time_value'),
            last_update_time=int(request.lastHistoryTime),
            request_user_name=request.cookies.get("requestUserName")
        )
        index = 0
        last_get_time = ''
        async for item in yield_response:
            if index == 0:
                last_get_time = item.viewInfo.viewTime
            index += 1
            yield item
        yield videoInfoResponse(
            errorCode=200,
            errorMsg=str(last_get_time),
            requestUserName=request.cookies.get("requestUserName"),
            webSiteName="bilibili",
        )
        end_time = time.time()
        print(f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 获取动态完毕，使用{request.cookies.get('requestUserName')}用户,时间参数是{request.lastHistoryTime}，耗时{int(end_time-start_time)}。获取到{index}个数据")

    
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
        print(f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 获取用户信息完毕，使用{request.cookies.get('requestUserName')}用户")
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
        print(f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} {client_ip} 按列表获取视频信息完毕，使用{request.cookies.get('requestUserName')}用户")
    
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


def create_server():
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    server_pb2_grpc.add_WebSiteServiceServicer_to_server(BilibiliServiceServicer(), server)
    server.add_insecure_port("[::]:50051")
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
    RegisterWebSite('bilibili', '192.168.0.20', 50051)
    try:
        asyncio.run(run_server(), debug=True)
    finally:
        InvalidGrpcClient('bilibili', '192.168.0.20', 50051)
