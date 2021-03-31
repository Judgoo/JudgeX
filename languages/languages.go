//go:generate go run ./generate

package languages

import "fmt"

/*
ENUM(
	Assembly asm
	Bash bash
	C c
	Clojure clj
	CoffeeScript coffee
	Cpp cpp
	CSharp cs
	D d
	Elixir ex
	Go go
	Groovy groovy
	Haskell hs
	Java java Main
	JavaScript js
	Julia jl
	Kotlin kt Main
	Lua lua
	Nim nim
	Ocaml ml
	Perl pl
	Php php
	Python py
	Raku raku
	Ruby rb
	Rust rust
	Scala scala
	Swift swift
	TypeScript ts
)
*/
type LanguageType int

type LanguageRecipe struct {
	Build []string
	Run   []string
}

// 可以把 version 传进来，然后决定返回不同的运行命令之类的。
func (lang LanguageType) Recipe() (LanguageRecipe, error) {
	fileName := lang.FileName()
	switch lang {
	case Assembly:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("nasm -f elf64 -o a.o %s", fileName),
				"ld -o a.out a.o",
			},
			[]string{
				"./a.out",
			},
		}, nil
	case Bash:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("bash %s", fileName),
			},
		}, nil
	case C:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("gcc -lm -w -O3 -std=gnu17 %s -o a.out", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case Clojure:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("clj -M %s", fileName),
			},
		}, nil
	case CoffeeScript:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("coffee %s", fileName),
			},
		}, nil
	case Cpp:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("g++ -lm -w -O3 -std=gnu++17 %s -o a.out", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case CSharp:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("mcs -out:a.exe %s", fileName),
			},
			[]string{"mono a.exe"},
		}, nil
	case D:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("dmd -ofa.out %s", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case Elixir:
		return LanguageRecipe{
			[]string{""},
			[]string{fmt.Sprintf("elixirc %s", fileName)},
		}, nil

	case Go:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("go build -o a.out %s", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case Groovy:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("groovy %s", fileName),
			},
		}, nil
	case Haskell:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("runghc %s", fileName),
			},
		}, nil
	case Java:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("javac %s", fileName),
			},
			[]string{
				// TODO 可以改成从 fileName 中读取
				fmt.Sprintf("java %s", "Main"),
			},
		}, nil
	case JavaScript:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("node %s", fileName),
			},
		}, nil
	case Julia:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("julia %s", fileName),
			},
		}, nil
	case Kotlin:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("kotlinc %s", fileName),
			},
			[]string{
				// TODO 可以改成从 fileName 中读取
				fmt.Sprintf("kotlin %sKt", "Main"),
			},
		}, nil
	case Lua:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("lua %s", fileName),
			},
		}, nil
	case Nim:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("nim --hints:off --verbosity:0 compile --run %s", fileName),
			},
		}, nil
	case Ocaml:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("ocamlc -o a.out %s", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case Perl:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("perl %s", fileName),
			},
		}, nil
	case Php:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("php %s", fileName),
			},
		}, nil
	case Python:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("python %s", fileName),
			},
		}, nil
	case Raku:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("raku %s", fileName),
			},
		}, nil
	case Ruby:
		return LanguageRecipe{
			[]string{""},
			[]string{
				fmt.Sprintf("ruby %s", fileName),
			},
		}, nil
	case Rust:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("rustc -o a.out %s", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case Scala:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("scalac %s", fileName),
			},
			[]string{"scala Main"},
		}, nil
	case Swift:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("swiftc -o a.out %s", fileName),
			},
			[]string{"./a.out"},
		}, nil
	case TypeScript:
		return LanguageRecipe{
			[]string{
				fmt.Sprintf("tsc -out a.js %s", fileName),
			},
			[]string{"node a.js"},
		}, nil
	default:
		return LanguageRecipe{}, fmt.Errorf("%s is not a valid LanguageType", lang)
	}
}

var OnlyCheckMem = []LanguageType{Java, Kotlin, JavaScript, TypeScript}
