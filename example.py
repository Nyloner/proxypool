# coding:utf-8
import requests

proxies = {
    'http': 'http://127.0.0.1:8080',
    'https': 'http://127.0.0.1:8080',
}

url = 'https://myip.ipip.net/'
html = requests.get(url, proxies=proxies,timeout=10).text
print(html)

