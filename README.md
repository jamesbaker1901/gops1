# gops1

`gops1` can be used to set a fancy bash prompt, aka `PS1`. It is not customizable, nor is it written particularly well or tested at all outside of my own daily use.

It looks like this:
```
┌─[jaybaker]─(~/go/src/github.com/jamesbaker1901/gops1) (master)
└─(tools|kube-system)[15:09]$ ps1min
~ ps1max
┌─[jaybaker]─(~/go/src/github.com/jamesbaker1901/gops1) (master)
└─(tools|kube-system)[15:10]$
```

In the past, I constructed my `PS1` using bash script as detailed [here](https://jay-baker.com/color-bash-prompt-ps1-with-git-integration/). It worked great but bash is slow and as the script got more and more complicated over the years it was starting to show. Since the `PROMPT_COMMAND` is executed every single time the `PS1` is drawn (i.e., every time you enter a command or even just hit enter), this delay can get _very_ annoying.

So I took my complicated `PROMPT_COMMAND` and rewrote it in go, more or less as is. It's much faster in go, but it's still not the best way to build your `PS1` (solutions like [powerline-shell](https://github.com/b-ryan/powerline-shell) are almost always a better solution for you.)

In any case, if you wish to build your own custom `PS1` in go you could use this project as a starting point. Likewise, if you happen to like the look of this one then feel free to use it! It works great for me.

## Features

`gops1` will show the current user, then the current working directory, followed by the current git branch (if any). The branch text will be green if there are no uncommitted changes and red if there are some.

It also supports reading your `~/.kube/config` file to get your current `kubectl` context and namespace if one is specified for that context. Additionally, if your `AWS_PROFILE` is set differently from your `kubectl` context it will display that difference as well:

```
┌─[jaybaker]─(~/go/src/github.com/jamesbaker1901/gops1) (master)
└─([a]dev[k]tools)[15:13]$
```

If you don't use `kubectl` and don't have a `~/.kube/config` file, the second line will simply show the time, like so:

```
┌─[jaybaker]─(~/go/src/github.com/jamesbaker1901/gops1) (master)
└─[15:13]$
```

`gops1` will display a green `$` prompt if the previous command exited with a 0 code, and a red `$` for other codes.

Finally, if you set the environment variable `GOPS1_MINIMAL=true` `gops1` will simply print a `~ `for a prompt, which is nice if you're finding all of the prompt info to be too cluttered. The prompt will stil show green/red depending on the exit code of the previous command though.

## Installation

``` 
go get github.com/jamesbaker1901/gops1
```

Then add this to you `~/.bashrc` or `~/.bash_profile` if you're on osx:

```
alias ps1min='export GOPS1_MINIMAL=true'
alias ps1max='unset GOPS1_MINIMAL'

# Set PS1
new_ps1 () {
        PS1="$(gops1 $?)"
}
PROMPT_COMMAND=new_ps1
```
The two aliases are optional, and only need to be added if you intend to occaisionally use the minimal mode (`~`).

Then do `source ~/.bashrc` (or `source ~/.bash_profile`) for it to take effect. Enjoy!

## Configuration

All configuration for `gops1` is done through environment variables. Supported variables are:

* `GOPS1_MINIMAL` | boolean | If true, `gops1` will return only a `~` character. Useful to declutter your terminal.
* `GOPS1_HOST` | boolean | If true, `gops1` will return a user and hostname for the first block, rather than the default of only user name. i.e., `[user@host]`
* `GOPS1_PWD_DEPTH` | integer | Max depth to display current working directory in the `PS1`. If you are 15 directories deep, the `PS1` will become very unweildy. This just replaces the first n directories with `...`, where n is current directory depth - `GOPS1_PWD_DEPTH`. Default is 10.

