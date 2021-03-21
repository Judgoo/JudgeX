# JudgeX

## 如何进行开发

首先需要安装好 go 1.16+

```bash
go mod download
```

### 调试

使用 [air](https://github.com/cosmtrek/air) 这个包进行调试。

安装好之后在当前目录执行 `air -c .air.toml` 就能调试了。

## 判题语言支持

所有支持的判题语言写在了 `./pkg/languages.go` 中：

```go
/* ENUM(
	Assembly asm
	...
)*/
type LanguageType int
```

需要增加的语言项按照这样写好，然后代码生成工具会帮我们生成具体的枚举类型和相关转换方法到 `./pkg/languages_enum.go` 这个文件。

如果需要增加语言项的话按格式添加到 ENUM 的括号中即可：

```go
/*
ENUM(
	<类型名> <文件后缀> <文件名（可不填，默认为 main）>
)
```

然后代码生成工具就会帮我们生成 `./pkg/languages_enum.go` 里面的内容。

这个文件的作用如下：

1. 从字符串解析生成语言项  
   比如说用户请求的是 /python 这个路径，我们只能拿到 `python` 这个字符串，如果我们能从 `python`、`Python` 等字符串转移到一个标准格式的话，在我们的代码中就很方便的标记用户请求的语言和进行相关的处理。
2. 方便的标记每个文件的文件名  
   因为不同的语言的代码文件的后缀不同，而且 Java, Kotlin 等语言还需要文件名首字母大写。所以一起在 `languages.go` 中枚举出来，后面可以直接用。

### 生成 `languages_enum` 文件

首先要安装我们魔改的 `go-enum`，帮我们生成这个文件。

```sh
mkdir -p ./third-party
git clone https://github.com/Judgoo/go-enum ./third-party/go-enum
cd ./third-party/go-enum && make build
```

然后想要生成的时候就在根目录下执行 `go generate ./pkg` 即可。

修改了 `go-enum` 的代码之后得重新 build `go-enum`，然后重新生成我们的代码。

## 注意事项

### commit 规范

1. 第一个单词需要使用动词且英文首字母大写  
   如：Fix/Add/Remove/Update
2. 第一行不要太长，简单描述即可  
   具体内容放在 commit 的 body 中
