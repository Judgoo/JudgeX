# JudgeX

## 如何进行开发

首先需要安装好 go 1.16+

```bash
go mod download
```

### 调试

使用 [air](https://github.com/cosmtrek/air) 进行调试。

安装好之后直接当前目录执行 `air` 就能调试了。

## 自动生成代码

`./pkg/languages_enum.go` 这个文件是靠代码生成工具生成的，如果需要增加语言项的话需要编辑 `./pkg/languages.go`：

```go
/*
ENUM(
	Assembly asm [main]
	代码类型枚举 文件后缀 文件名，可不填
	...
)
```

然后代码生成工具就会帮我们生成 `./pkg/languages_enum.go`

首先要安装我们自己维护的 `go-enum` (方便生成我们自己需要的代码)。

```sh
mkdir -p ./third-party
git clone https://github.com/Judgoo/go-enum ./third-party/go-enum
cd ./third-party/go-enum && make build
```

然后想要生成 `languages_enum.go` 的时候就在根目录下执行 `go generate ./pkg`。

如果你修改了 `go-enum` 的话记得重新 build。
