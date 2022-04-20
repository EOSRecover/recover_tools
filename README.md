# recover_tools

*what is this?*

recover-tools is simple send proposal to change hacker's eos account auth by BPs multisignature

## How to use?

1. Copy conf/conf.example.yaml to conf.yaml
2. Edit conf.yaml content to needs info
3. Use go run to send proposal or go build to get execute file

### Use go run
`go run main.go proposal`

### Use go build
```
go build -a -installsuffix cgo -ldflags '-s -w' -o tools *.go
./tools proposal
````
