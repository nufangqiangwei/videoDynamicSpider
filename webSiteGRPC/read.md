## 备注
为了省事，所有的网站请求方法都改成grpc方式调用。爬虫端只负责定时调用爬取  
爬虫端拿到实现了models.VideoCollection这个接口的对象，调用这个获取数据报错到数据库  
现在需要将这些网站的请求的代码拆分出去，这里不在写了。这里只负责数据处理的逻辑
