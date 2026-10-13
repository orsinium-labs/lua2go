# lua2go

A simple CLI tool for converting code from Lua to Go syntax.

It only handles syntax (and even that still can be a bit off). You still need to manually replace `any` with proper types, implement or import missing functions, run `go fmt`.

The main purpose of the project is to assist in rewriting code from Lua to Go, not to fully automate it. I use it for porting games from [PICO-8](https://www.lexaloffle.com/pico-8.php) to [Firefly Zero](https://fireflyzero.com/), often with an assistance of [pico8dis](https://github.com/orsinium-labs/pico8dis/) (and always with the author's permission, of course).

## Installation

```bash
go install github.com/orsinium-labs/lua2go@main
```

## Usage

```bash
cat main.lua | lua2go > main.go
```
