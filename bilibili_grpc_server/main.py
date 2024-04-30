# todo 生成GRPC接口命令
# python -m grpc_tools.protoc -I E:\PythonCode\bilibili_grpc_server --python_out=. --pyi_out=. --grpc_python_out=. E:\PythonCode\bilibili_grpc_server\server.proto
import asyncio

import grpc
from concurrent import futures

from bilibili_api import Credential, video, hot
from bilibili_api.user import User

import server_pb2_grpc

from bili import get_self_user_dynamic, get_self_user_view_history
from server_pb2 import AuthorInfoResponse, videoInfoResponse, classifyInfoResponse


class BilibiliServiceServicer(server_pb2_grpc.WebSiteServiceServicer):
    async def GetUserFollowUpdate(self, request, context):
        print(request)
        yield_response = get_self_user_dynamic(
            request.cookies.get("sessdata"),
            request.cookies.get('bili_jct'),
            request.cookies.get('buvid3'),
            request.cookies.get('dedeuserid'),
            request.cookies.get('ac_time_value')
        )

        async for item in yield_response:
            yield item

    async def GetUserViewHistory(self, request, context):
        print(request)
        yield_response = get_self_user_view_history(
            request.cookies.get("sessdata"),
            request.cookies.get('bili_jct'),
            request.cookies.get('buvid3'),
            request.cookies.get('dedeuserid'),
            request.cookies.get('ac_time_value')
        )

        async for item in yield_response:
            yield item

    async def GetSelfInfo(self, request, context):
        print(request)
        credential = Credential(
            sessdata=request.userInfo.cookies.get("sessdata"),
            bili_jct=request.userInfo.cookies.get('bili_jct'),
            buvid3=request.userInfo.cookies.get('buvid3'),
            dedeuserid=request.userInfo.cookies.get('dedeuserid'),
            ac_time_value=request.userInfo.cookies.get('ac_time_value'),
        )
        user = User(request.cookies.get('dedeuserid'), credential=credential)
        user_info = await user.get_user_info()

        return AuthorInfoResponse(
            name=user_info.get('name'),
            avatar=user_info.get('face'),
            uid=str(user_info.get('mid')),
            desc=user_info.get('sign'),
            followNumber=user_info.get('following'),
        )

    async def GetVideoList(self, request, context):
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
            print('bvid: ', wait_video)
            yield await self.get_video_list(credential, wait_video)

    async def GetHotVideoList(self, request, context):
        print(request)
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
    await bilibili_server.wait_for_termination()

if __name__ == '__main__':
    asyncio.run(run_server())
