
# ProxyPool

A simple proxy tool written by golang.

## Usage

Config

```json
{
  "ServerIP": "localhost",
  "ServerPort": 8080, // 监听
  "ProxyRedis": {
    "Server": "127.0.0.1:6379" // redis服务地址，依赖redis zset存储代理IP
  }
}
```

Run Server

```Python
make run
```

Example

```Python
import requests

proxies = {
    'http': 'http://127.0.0.1:8080',
    'https': 'http://127.0.0.1:8080',
}

url = 'https://myip.ipip.net/'
resp = requests.get(url, proxies=proxies,timeout=10).text
print(resp)
```
