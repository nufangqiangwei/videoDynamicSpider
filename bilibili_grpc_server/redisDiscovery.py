import redis

r = redis.Redis(host='localhost', port=6379, db=5)


def RegisterWebSite(web_site_name, ip, port):
    r.sadd("grpcServerType", web_site_name)
    r.sadd(web_site_name, f'{ip}:{port}')
    r.publish('grpcServerTypeSub', f'register-bilibili-{ip}:{port}')


def InvalidGrpcClient(web_site_name, ip, port):
    r.srem(web_site_name, f'{ip}:{port}')
    if r.scard(web_site_name) == 0:
        r.srem("grpcServerType", web_site_name)
    r.publish('grpcServerTypeSub', f'unregister-bilibili-{ip}:{port}')
