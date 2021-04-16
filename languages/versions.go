package languages

// 每一个版本的描述
type VersionInfo struct {
	ImageName   string
	Description string
	ExampleCode string
}

// 名称 -> 描述
type Versions map[string]VersionInfo

var VersionMap = map[LanguageType]Versions{
	Assembly: {
		"nasm": {
			ImageName:   "judgoo/nasm:v0.0.1",
			Description: "nasm on alpine",
			ExampleCode: `section .data
    msg db "helloworld", 0ah

section .text
    global _start
_start:
    mov rax, 1
    mov rdi, 1
    mov rsi, msg
    mov rdx, 10
    syscall
    mov rax, 60
    mov rdi, 0
    syscall
`,
		},
	},
	Bash: {
		"bash": {
			ImageName:   "judgoo/bash:v0.0.1",
			Description: "bash on alpine",
			ExampleCode: `echo "helloworld"`,
		},
	},
	C: {
		"gcc8": {
			ImageName:   "judgoo/gpp:v0.0.1",
			Description: "gcc on debian",
			ExampleCode: `#include <stdio.h>
int main()
{
  int a, b;
  while (scanf("%d %d", &a, &b) != EOF)
    printf("%d\n", a + b);
  return 0;
}`,
		},
	},
	CSharp: {
		"csharp": {
			ImageName:   "judgoo/csharp:v0.0.1",
			Description: "csharp on alpine",
			ExampleCode: `using System;
using System.Collections.Generic;
using System.Linq;

class MainClass {
    static void Main() {
        Console.WriteLine("helloworld");
    }
}`,
		},
	},
	Cpp: {
		"g++8": {
			ImageName:   "judgoo/gpp:v0.0.1",
			Description: "g++ in debian",
			ExampleCode: `#include <iostream>
using namespace std;

int main()
{
    cout << "helloworld" << endl;
    return 0;
}`,
		},
	},
	D: {
		"dlang2": {
			ImageName:   "judgoo/dlang2:v0.0.1",
			Description: "dlang2 on alpine",
			ExampleCode: `import std.stdio;

void main()
{
    writeln("helloworld");
}`,
		},
	},
	Go: {
		"golang16": {
			ImageName:   "judgoo/golang:v0.0.1",
			Description: "golang on alpine",
			ExampleCode: `package main

import (
	"fmt"
)

func main() {
	fmt.Println("helloworld")
}`,
		},
	},
	Haskell: {
		"nasm": {
			ImageName:   "judgoo/haskell:v0.0.1",
			Description: "haskell on alpine",
			ExampleCode: `main = putStrLn "Hello World!"`,
		},
	},
	Java: {
		"openjdk8": {
			ImageName:   "judgoo/kotlin1.42:v0.0.1",
			Description: "openjdk8 on debian",
			ExampleCode: `class Main {
    public static void main(String[] args) {
        System.out.println("helloworld");
    }
}`,
		},
		"openjdk11": {
			ImageName:   "judgoo/openjdk11:v0.0.1",
			Description: "openjdk11 on debian",
			ExampleCode: `class Main {
    public static void main(String[] args) {
        System.out.println("helloworld");
    }
}`,
		},
	},
	JavaScript: {
		"nodejs14": {
			ImageName:   "judgoo/nodejs14:v0.0.1",
			Description: "nodejs14 on alpine",
			ExampleCode: `console.log('helloworld');`,
		},
	},
	Julia: {
		"julia1.6": {
			ImageName:   "judgoo/julia1.6:v0.0.1",
			Description: "julia on alpine",
			ExampleCode: `println("helloworld")`,
		},
	},
	Kotlin: {
		"kotlin1.42": {
			ImageName:   "judgoo/kotlin1.42:v0.0.1",
			Description: "kotlin1.42 on debian",
			ExampleCode: `fun main(args: Array<String>){
    println("helloworld")
}`,
		},
	},
	Lua: {
		"lua": {
			ImageName:   "judgoo/lua:v0.0.1",
			Description: "lua on alpine",
			ExampleCode: `print("helloworld");`,
		},
	},
	Ocaml: {
		"ocaml": {
			ImageName:   "judgoo/ocaml:v0.0.1",
			Description: "ocaml on alpine",
			ExampleCode: `print_endline "helloworld"`,
		},
	},
	Perl: {
		"perl": {
			ImageName:   "judgoo/perl:v0.0.1",
			Description: "perl on alpine",
			ExampleCode: `print "helloworld\n";`,
		},
	},
	Php: {
		"php": {
			ImageName:   "judgoo/php:v0.0.1",
			Description: "php on alpine",
			ExampleCode: `<?php
    echo "helloworld\n";`,
		},
	},
	Python: {
		"python3.9": {
			ImageName:   "judgoo/python3.9:v0.0.1",
			Description: "python3.9 on debian",
			ExampleCode: `print("helloworld");`,
		},
		"python3.9w": {
			ImageName:   "judgoo/python3.9w:v0.0.1",
			Description: "python3.9 with numpy and pandas on debian",
			ExampleCode: `print("helloworld");`,
		},
	},
	Ruby: {
		"ruby": {
			ImageName:   "judgoo/ruby:v0.0.1",
			Description: "ruby on alpine",
			ExampleCode: `puts "helloworld"`,
		},
	},
	Rust: {
		"rust1.51": {
			ImageName:   "judgoo/rust1.51:v0.0.1",
			Description: "rust1.51 on debian",
			ExampleCode: `fn main() {
    println!("helloworld");
}`,
		},
	},
	Scala: {
		"scala": {
			ImageName:   "judgoo/scala:v0.0.1",
			Description: "scala on debian",
			ExampleCode: `object Main extends App {
    println("helloworld")
}`,
		},
	},
	Swift: {
		"swift": {
			ImageName:   "judgoo/swift:v0.0.1",
			Description: "swift on debian",
			ExampleCode: `print("helloworld")`,
		},
	},
	TypeScript: {
		"typescript": {
			ImageName:   "judgoo/typescript:v0.0.1",
			Description: "typescript on alpine",
			ExampleCode: `console.log('helloworld');`,
		},
	},
}

func (lang *LanguageType) GetVersions(version string) Versions {
	return VersionMap[*lang]
}

func (lang *LanguageType) GetVersion(version string) (string, VersionInfo, bool) {
	var (
		vInfo VersionInfo
		vName string
		ok    bool
	)
	if version == "" {
		// 获取第一个 version
		for name, v := range VersionMap[*lang] {
			vName = name
			vInfo = v
			break
		}
		return vName, vInfo, true
	}
	vInfo, ok = VersionMap[*lang][version]
	return version, vInfo, ok
}
