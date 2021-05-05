package languages

var ExamplesMap = map[LanguageType]string{
	Assembly: `section .data
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
	Bash: `echo "helloworld"`,
	C: `#include <stdio.h>
int main()
{
  int a, b;
  while (scanf("%d %d", &a, &b) != EOF)
    printf("%d\n", a + b);
  return 0;
}`,
	CSharp: `using System;
using System.Collections.Generic;
using System.Linq;

class MainClass {
    static void Main() {
        Console.WriteLine("helloworld");
    }
}`,
	Cpp: `#include <iostream>
using namespace std;

int main()
{
    cout << "helloworld" << endl;
    return 0;
}`,
	D: `import std.stdio;

void main()
{
    writeln("helloworld");
}`,
	Go: `package main

import (
	"fmt"
)

func main() {
	fmt.Println("helloworld")
}`,
	Haskell: `main = putStrLn "Hello World!"`,
	Java: `class Main {
    public static void main(String[] args) {
        System.out.println("helloworld");
    }
}`,
	JavaScript: `console.log('helloworld');`,
	Julia:      `println("helloworld")`,
	Kotlin: `fun main(args: Array<String>){
    println("helloworld")
}`,
	Lua:   `print("helloworld");`,
	Ocaml: `print_endline "helloworld"`,
	Perl:  `print "helloworld\n";`,
	Php: `<?php
    echo "helloworld\n";`,
	Python: `print("helloworld");`,
	Ruby:   `puts "helloworld"`,
	Rust: `fn main() {
    println!("helloworld");
}`,
	Scala: `object Main extends App {
    println("helloworld")
}`,
	Swift:      `print("helloworld")`,
	TypeScript: `console.log('helloworld');`,
}
