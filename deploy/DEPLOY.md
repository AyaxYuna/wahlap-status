# 上线部署清单(maimai.imikufans.cn)

## 0. 准备

服务器需安装 Go ≥1.23、Node ≥20、nginx、certbot。域名已解析到本机。

## 1. 拉代码

```bash
git clone https://github.com/shenwenbo2009/maimaistatus.git
cd maimaistatus
```

## 2. 后端

```bash
cd backend
go build -o gateway .
# 环境变量(写入 /etc/maimai-gateway.env,权限 600):
#   PORT=8080
#   LXNS_CLIENT_ID=86a5bacf-f8a1-4459-8687-5702fdf2cf27
#   LXNS_CLIENT_SECRET=你的应用密钥
#   LXNS_REDIRECT_URI=https://maimai.imikufans.cn/lxns-fallback
#   LXNS_SCOPES=read_user_profile write_player read_player
```

systemd(见 `deploy/maimai-gateway.service`,改 `User=` 与 `WorkingDirectory=` 后):

```bash
sudo cp deploy/maimai-gateway.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now maimai-gateway
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/api/lxns/auth-url  # 应返回生产回调的授权链接
```

## 3. 前端

```bash
cd ../frontend
npm install
NUXT_PUBLIC_API_BASE=https://maimai.imikufans.cn/api npm run build
```

用 pm2 或 systemd 常驻 `.output/server/index.mjs`(端口 3000):

```bash
node .output/server/index.mjs  # 默认 3000,可用 PORT 覆盖
```

## 4. 反代与证书

nginx 见 `deploy/nginx.conf`(改域名与本机端口),要点:

- `/` → 127.0.0.1:3000(Nuxt SSR)
- `/api/` → 127.0.0.1:8080,注意 `/api/stream` 必须 `proxy_buffering off`(SSE)
- certbot 申请证书后重载 nginx

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-enabled/maimai
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d maimai.imikufans.cn
```

## 5. 验证

- `https://maimai.imikufans.cn/` 首页状态正常
- 查分 → 登录 → 落雪授权 → 跳回 `/lxns-fallback` → 显示玩家卡与 B50
- `https://maimai.imikufans.cn/api/healthz` → `{"ok":true,...}`

## 6. 落雪面板确认

回调地址包含 `https://maimai.imikufans.cn/lxns-fallback`(与网关 `LXNS_REDIRECT_URI` 字符级一致)。
