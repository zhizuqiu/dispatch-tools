# dispatch-up

```
dispatch-up -f ./some.zip
dispatch-up -a http://127.0.0.1:8080/ -f ./some.zip
dispatch-up -a http://127.0.0.1:8080/ -d /temp/ -f ./some.zip
```

### build

```
GOOS=linux GOARCH=amd64 go build -o dist/amd64/dispatch-up
```