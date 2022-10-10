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

### build

amd64:
```bash
GOOS=linux GOARCH=amd64 go build -o dist/amd64/dispatch
```

arm64:
```bash
GOOS=linux GOARCH=arm64 go build -o dist/arm64/dispatch
```

ppc64le:
```bash
GOOS=linux GOARCH=ppc64le go build -o dist/ppc64le/dispatch
```