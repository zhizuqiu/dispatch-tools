# dispatch-tools

```
dispatch tools

Usage:
  dispatch [command]

Available Commands:
  config      查看配置文件
  download    下载文件
  help        Help about any command
  list        查询文件
  server      启动文件服务端
  up          上传文件

Flags:
  -a, --address string         dispatch server 的地址，例如：-a http://127.0.0.1:8080/
      --config string          配置文件 (默认：$HOME/.dispatch/dispatch.yaml)
  -d, --dir string             dispatch server 的目录路径，例如：-d /temp/
  -h, --help                   help for dispatch
      --http-password string   dispatch server 的认证密码
      --http-user string       dispatch server 的认证用户

Use "dispatch [command] --help" for more information about a command.
```

### 服务端

单个二进制即可启动一个与 Java 版（dispatch-java）协议兼容的文件服务端：

```bash
# 默认：数据目录 ./data，监听 127.0.0.1:8080，无认证
dispatch server

# 指定数据目录、对外监听、开启认证
dispatch server -d /data --listen 0.0.0.0:8080 --http-user admin --http-password 123456

# 同名文件已存在时返回 409，不覆盖
dispatch server -d /data --no-overwrite
```

启动后，客户端可直接上传、查看、下载：

```bash
dispatch up -a http://<host>:8080/ -d /temp/ ./some.zip
dispatch list -a http://<host>:8080/ -d /temp/
dispatch download -a http://<host>:8080/ -p ./ http://<host>:8080/data/temp/some.zip
```

| Flag | 说明 | 默认 |
|---|---|---|
| `-d, --dir` | 本地数据根目录（不存在则自动创建） | `./data` |
| `--listen` | 监听地址 | `127.0.0.1:8080` |
| `--http-user` / `--http-password` | Basic Auth，两者都为空则关闭 | 空 |
| `--no-overwrite` | 上传同名文件时返回 409 | 关闭 |

接口与 Java 版保持一致：`/file/List`、`/file/UploadFile`、`/file/newDir`、`/file/Delete`、`/data/...`，两个服务端可以互换。

认证账号也可以直接写在配置文件（默认 `~/.dispatch/dispatch.yaml`）的 `http-user` / `http-password` 中，服务端和客户端会读取同一份配置。

### build

```bash
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64   go build -o dist/amd64/dispatch
CGO_ENABLED=0 GOOS=linux   GOARCH=arm64   go build -o dist/arm64/dispatch
CGO_ENABLED=0 GOOS=linux   GOARCH=ppc64le go build -o dist/ppc64le/dispatch
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64   go build -o dist/darwin-amd64/dispatch
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64   go build -o dist/darwin-arm64/dispatch
CGO_ENABLED=0 GOOS=windows GOARCH=amd64   go build -o dist/windows-amd64/dispatch.exe
```

