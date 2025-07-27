import requests

def get_public_ip_ipinfo():
    """通过 ipinfo.io 获取公网 IP 地址"""
    try:
        response = requests.get('https://ipinfo.io/ip') # 直接获取 IP 地址
        if response.status_code == 200:
            return response.text.strip()
        else:
            print(f"请求失败，状态码: {response.status_code}")
            return None
    except requests.exceptions.RequestException as e:
        print(f"网络请求错误: {e}")
        return None

def get_public_ip_json_full():
    """通过 ipinfo.io 获取包含 IP 在内的完整 JSON 信息"""
    try:
        response = requests.get('https://ipinfo.io/json')
        if response.status_code == 200:
            data = response.json()
            return data.get('ip') # 从 JSON 中提取 'ip' 字段
        else:
            print(f"请求失败，状态码: {response.status_code}")
            return None
    except requests.exceptions.RequestException as e:
        print(f"网络请求错误: {e}")
        return None

if __name__ == "__main__":
    print("尝试获取公网 IP (简单方式):")
    public_ip = get_public_ip_ipinfo()
    if public_ip:
        print(f"你的公网 IP 地址是: {public_ip}")
    else:
        print("无法获取公网 IP 地址。")

    print("\n尝试获取公网 IP (从完整 JSON):")
    public_ip_json = get_public_ip_json_full()
    if public_ip_json:
        print(f"你的公网 IP 地址是: {public_ip_json}")
    else:
        print("无法从 JSON 获取公网 IP 地址。")