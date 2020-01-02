# vargs

Open given files / Call vim function with given strings using Terminal API (xargs for vim)

## Usage

```
vargs [-0] [-s {separators (comma-separated)}] [-I {replstr}] [-t] [{vim terminal-api arguments (default: "drop")} ...]
```

`vargs` sends given arguments to Vim using Terminal API.<br>
If no arguments were given, `"drop"` is specified implicitly (default).

See below about Vim Terminal API.

* [Vim help (English)](https://vim-jp.org/vimdoc-en/terminal.html#terminal-api)
* [Vim help (Japanese)](https://vim-jp.org/vimdoc-ja/terminal.html#terminal-api)

#### Options

```
  -0    Change separator to NUL character. This is same as "-s nul"
  -I string
        If this replacement string was given, replace arguments by this with each item
  -s string
        Change separators with these comma-separated values (available values are "space", "tab", "newline", "nul") (default "newline")
  -t    Print JSON command to standard error before printing with escape sequence (verbose print)
```

## Why you created this?

I wanted a way to quickly open selected file using file fuzzy finder UI, like [gof](https://github.com/mattn/gof).<br>
I sent [a pull request](https://github.com/mattn/gof/pull/14) to gof that adds `-t` option to open selected file using Vim Terminal API.<br>
After that, I got an idea that creating generic command for Vim Terminal API.

Vim Terminal API is very handy to communicate from processes inside Vim terminal window to outer Vim.

Other Vim plugin examples using Terminal API:

* [sync-term-cwd.vim](https://github.com/tyru/sync-term-cwd.vim)
  * Sync Vim current directory with current directory of bash/zsh
* [tapi-reg.vim](https://github.com/tyru/tapi-reg.vim)
  * Access Vim clipboard from bash/zsh


## Examples

### Open the selected file (peco, gof)

Using peco

```
$ ls | peco | vargs
```

Using gof (recurse subdirectories)

```
$ gof | vargs
```

However, [`gof` already has `-t` option](https://github.com/mattn/gof/pull/14) to open in Vim :smirk:

```
$ gof -t
```

### Open all files under repository

```
$ find . -path ./.git -prune -o -type f | vargs
```

More safe way...

```
$ find . -path ./.git -prune -o -type f -print0 | vargs -0
```

### Open project under $GOPATH

```
$ ls -d $GOPATH/src/github.com/*/* | peco | vargs
```

You may prefer [project-guide.vim](https://github.com/tyru/project-guide.vim).
