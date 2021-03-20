//go:generate go-enum -f=$GOFILE --marshal --nocase --noprefix

package pkg

import "fmt"

/*
ENUM(
	Assembly
	Bash
	C
	Clojure
	CoffeeScript
	Cpp
	Csharp
	D
	Elixir
	Erlang
	Fsharp
	Go
	Groovy
	Haskell
	Java
	JavaScript
	Julia
	Kotlin
	Lua
	Mercury
	Nim
	Ocaml
	Perl
	Php
	Python
	Raku
	Ruby
	Rust
	Scala
	Swift
	TypeScript
)
*/
type LanguageType int

type LanguageRecipe struct {
	Build []string
	Run   []string
}

func (lang LanguageType) Recipe(files []string) (LanguageRecipe, error) {
	if len(files) == 0 {
		return LanguageRecipe{}, fmt.Errorf("please check source files")
	}

	switch lang {
	case Assembly:
		return LanguageRecipe{
			[]string{},
			[]string{},
		}, nil
	case Bash:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case C:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Clojure:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case CoffeeScript:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Cpp:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Csharp:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case D:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Elixir:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Erlang:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Fsharp:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Go:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Groovy:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Haskell:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Java:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case JavaScript:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Julia:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Kotlin:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Lua:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Mercury:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Nim:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Ocaml:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Perl:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Php:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Python:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Raku:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Ruby:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Rust:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Scala:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case Swift:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	case TypeScript:
		return LanguageRecipe{
			[]string{""},
			[]string{""},
		}, nil
	default:
		return LanguageRecipe{}, fmt.Errorf("%s is not a valid LanguageType", lang)
	}
}
