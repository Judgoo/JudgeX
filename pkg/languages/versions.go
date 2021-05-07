package languages

// 每一个版本的描述
type VersionInfo struct {
	DisplayName string
	ImageName   string
	Description string
}

var VersionInfos = map[string]*VersionInfo{
	"nasm": {
		DisplayName: "NASM 2.15.05",
		ImageName:   "judgoo/nasm:v0.0.1",
		Description: "Assembly(2.15.05)",
	},
	"bash": {
		DisplayName: "5.1.0",
		ImageName:   "judgoo/bash:v0.0.1",
		Description: "Bash(5.1.0)",
	},
	"gcc8": {
		DisplayName: "GCC 8.3.0",
		ImageName:   "judgoo/gpp:v0.0.1",
		Description: "debian (GCC 8.3.0-6)",
	},
	"csharp": {
		DisplayName: "Mono 6.12.0.122",
		ImageName:   "judgoo/csharp:v0.0.1",
		Description: "CSharp(Mono 6.12.0.122)",
	},
	"dlang2": {
		DisplayName: "DMD v2.095.0",
		ImageName:   "judgoo/dlang2:v0.0.1",
		Description: "dlang2 on alpine",
	},
	"golang1.16": {
		DisplayName: "Go 1.16.3",
		ImageName:   "judgoo/golang:v0.0.1",
		Description: "golang on alpine",
	},
	"haskell": {
		DisplayName: "GHC 8.8.4",
		ImageName:   "judgoo/haskell:v0.0.1",
		Description: "haskell on alpine",
	},
	"openjdk8": {
		DisplayName: "OpenJDK 1.8.0",
		ImageName:   "judgoo/kotlin1.4.32:v0.0.1",
		Description: "openjdk8 on debian",
	},
	"openjdk11": {
		DisplayName: "OpenJDK 11.0.10",
		ImageName:   "judgoo/openjdk11:v0.0.1",
		Description: "openjdk11 on debian",
	},
	"nodejs14": {
		DisplayName: "Node.js 14.16.1",
		ImageName:   "judgoo/nodejs14:v0.0.1",
		Description: "nodejs14 on alpine",
	},
	"julia1.6": {
		DisplayName: "Julia 1.6.0",
		ImageName:   "judgoo/julia1.6:v0.0.1",
		Description: "julia on alpine",
	},
	"kotlin1.4": {
		DisplayName: "Kotlin 1.4.32",
		ImageName:   "judgoo/kotlin1.4.32:v0.0.1",
		Description: "kotlin1.42 on debian",
	},
	"lua": {
		DisplayName: "Lua 5.1.5",
		ImageName:   "judgoo/lua:v0.0.1",
		Description: "lua on alpine",
	},
	"ocaml": {
		DisplayName: "Ocaml 4.12.0",
		ImageName:   "judgoo/ocaml:v0.0.1",
		Description: "ocaml on alpine",
	},
	"perl": {
		DisplayName: "Perl 5.32.0",
		ImageName:   "judgoo/perl:v0.0.1",
		Description: "perl on alpine",
	},
	"php7": {
		DisplayName: "PHP 7.4.15",
		ImageName:   "judgoo/php7:v0.0.1",
		Description: "php on alpine",
	},
	"python3.9": {
		DisplayName: "Python 3.9.4",
		ImageName:   "judgoo/python3.9:v0.0.1",
		Description: "python3.9 on debian",
	},
	"python2.7": {
		DisplayName: "Python 2.7.18",
		ImageName:   "judgoo/python2.7:v0.0.1",
		Description: "python3.9 with numpy and pandas on debian",
	},
	"ruby": {
		DisplayName: "Ruby 2.7.3",
		ImageName:   "judgoo/ruby:v0.0.1",
		Description: "ruby on alpine",
	},
	"rust1.51": {
		DisplayName: "Rust 1.51.0",
		ImageName:   "judgoo/rust1.51:v0.0.1",
		Description: "rust1.51 on debian",
	},
	"scala2.13": {
		DisplayName: "Scala 2.13.5",
		ImageName:   "judgoo/scala2.13.5:v0.0.1",
		Description: "scala on debian",
	},
	"swift5.3": {
		DisplayName: "Swift 5.3.3",
		ImageName:   "judgoo/swift5.3.3:v0.0.1",
		Description: "swift on debian",
	},
	"typescript": {
		DisplayName: "esbuild 0.11.14",
		ImageName:   "judgoo/typescript:v0.0.1",
		Description: "typescript on alpine",
	},
}

var VersionNameMap = map[LanguageType][]string{
	Assembly:   {"nasm"},
	Bash:       {"bash"},
	C:          {"gcc8"},
	CSharp:     {"csharp"},
	Cpp:        {"gcc8"},
	D:          {"dlang2"},
	Go:         {"golang1.16"},
	Haskell:    {"haskell"},
	Java:       {"openjdk8", "openjdk11"},
	JavaScript: {"nodejs14"},
	Julia:      {"julia1.6"},
	Kotlin:     {"kotlin1.4"},
	Lua:        {"lua"},
	Ocaml:      {"ocaml"},
	Perl:       {"perl"},
	Php:        {"php7"},
	Python:     {"python3.9", "python2.7"},
	Ruby:       {"ruby"},
	Rust:       {"rust1.51"},
	Scala:      {"scala2.13"},
	Swift:      {"swift5.3"},
	TypeScript: {"typescript"},
}

func (lt *LanguageType) GetVersionNames(version string) []string {
	return VersionNameMap[*lt]
}

func (lt *LanguageType) GetVersionInfo(version string) (string, *VersionInfo, bool) {
	var (
		vInfo *VersionInfo
		vName string = version
		ok    bool
	)
	if version == "" {
		versions := VersionNameMap[*lt]
		vName = versions[0]
		vInfo = VersionInfos[vName]
		ok = true
	} else {
		vInfo, ok = VersionInfos[version]
	}
	return vName, vInfo, ok
}
