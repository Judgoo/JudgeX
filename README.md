# JudgeX

## 如何进行开发

首先需要安装好 go 1.16+

```bash
go mod download
```

### 运行

使用 [air](https://github.com/cosmtrek/air) 这个包运行代码。

安装好之后在当前目录执行 `air` 就能运行起来了。

你可以创建一个 `.env` 文件：

```dotenv
SERVER_PORT="8001"
SERVER_HOST="0.0.0.0"
```

里面填上需要的信息。

## 判题语言支持

所有支持的判题语言写在了 `./languages/languages.yml` 中：

```go
type LanguageType int
```

需要增加的语言项按照这样写好，然后代码生成工具会帮我们生成具体的类型到 `./languages/languages_impl.go` 和 `./languages/languages_impl.yml` 这两个文件。

如果需要增加语言项的话按格式添加到 `./languages/languages.yml` 中即可：

### 生成 `languages_impl` 文件

然后想要生成的时候就在根目录下执行 `go generate ./languages` 即可。

## 容器支持

可以安装 crun 作为运行时。

### 设置 Docker 默认运行时为 crun

编辑 /etc/docker/daemon.jso：

```json{
{
  "default-runtime": "crun",
  "runtimes": {
    "crun": {
      "path": "/usr/local/crun"
    }
  }
}
```

然后重启 docker。

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

## 注意事项

### commit 规范

1. 第一个单词需要使用动词且英文首字母大写  
   如：Fix/Add/Remove/Update
2. 第一行不要太长，简单描述即可  
   具体内容放在 commit 的 body 中

## 软件包

```bash
sudo apt-get install libdevmapper-dev
sudo apt-get install libbtrfs-dev
sudo apt-get install btrfs-progs

export PKG_CONFIG_PATH="/usr/lib/x86_64-linux-gnu/pkgconfig/"
```
