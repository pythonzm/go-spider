# go-spider
爬取某你懂的动图网站的出处动图，并生成命令行工具

直接使用go build spider.go 生成可执行文件，在Windows下生成的文件会比较大，大概7M多，可以使用 https://github.com/upx/upx 对生成的spider.exe进行压缩 `upx.exe -9 -k spider.exe`

可以使用 `spider -h` 查看支持的参数，Windows下是`spider.exe -h` 

```
[root@test spider]# spider -h
NAME:
   spider - 爬取某动图网站的出处动图

USAGE:
   spider [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --num value, -n value   设置默认下载并发数量，默认是10 (default: 10)
   --page value, -p value  设置下载的页数，默认是2 (default: 2)
   --to PATH               设置下载路径，windows默认路径是D:\gifs,其他系统默认是/tmp/gifs PATH (default: "/tmp/gifs/")
   --help, -h              show help
   --version, -v           print the version
```
