package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"
)

var Dir string
var ExitCode string

type Renderer interface {
	Render() (int, string)
}

type Config struct {
	User          User          `yaml:"user,omitempty"`
	Pwd           Pwd           `yaml:"pwd,omitempty"`
	Line          Line          `yaml:"line,omitempty"`
	Space         Space         `yaml:"space,omitempty"`
	Git           Git           `yaml:"git,omitempty"`
	NewLine       NewLine       `yaml:"newLine,omitempty"`
	TopBracket    TopBracket    `yaml:"topBracket,omitempty"`
	BottomBracket BottomBracket `yaml:"bottomBracket,omitempty"`
	Context       Context       `yaml:"context,omitempty"`
	Time          Time          `yaml:"time,omitempty"`
	Prompt        Prompt        `yaml:"prompt,omitempty"`
}

type Block struct {
	Formatting string `yaml:"formatting"`
	Prefix     string `yaml:"prefix"`
	Suffix     string `yaml:"suffix"`
	Content    string `yaml:"content"`
}

type User struct {
	Order int   `yaml:"order"`
	Host  bool  `yaml:"host,omitempty"`
	Block Block `yaml:"block,omitempty"`
}

type Pwd struct {
	Order    int   `yaml:"order"`
	MaxDepth int   `yaml:"maxDepth,omitempty"`
	Block    Block `yaml:"block,omitempty"`
}

type Git struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type Line struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type Space struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type NewLine struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type TopBracket struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type BottomBracket struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type Context struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type Time struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

type Prompt struct {
	Order int   `yaml:"order"`
	Block Block `yaml:"block,omitempty"`
}

func init() {
	Dir, _ = os.Getwd()
	ExitCode = "0"
	if len(os.Args) >= 2 {
		ExitCode = os.Args[1]
	}
}

func main() {
	ps1 := make([]string, 15)
	var c Config
	//c := yaml.MapSlice{}
	bytes, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		fmt.Println(err)
	}

	v := reflect.ValueOf(c)

	//fmt.Println(c)

	/*
		for _, v := range c {
			fmt.Println(v.Key)
			b := []byte(v.Value.(string))
			switch v.Key {
			case "topBracket":
				var topBracket TopBracket
				err = yaml.Unmarshal(b, &topBracket)
				if err != nil {
					fmt.Println(err)
				}
			}
			fmt.Println(v.Value)
		}
	*/

	values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
		element := values[i].(Renderer)
		order, rendered := element.Render()
		ps1[order] = rendered
	}

	ps1 = append(ps1, `\[\033[00m\] `)

	var out strings.Builder
	for _, val := range ps1 {
		out.WriteString(val)
	}

	fmt.Println(out.String())

}

func (s Space) Render() (int, string) {
	return s.Order - 1, s.Block.Render()
}

func (l Line) Render() (int, string) {
	return l.Order - 1, l.Block.Render()
}

func (n NewLine) Render() (int, string) {
	return n.Order - 1, n.Block.Render()
}

func (t TopBracket) Render() (int, string) {
	return t.Order - 1, t.Block.Render()
}

func (b BottomBracket) Render() (int, string) {
	return b.Order - 1, b.Block.Render()
}

func (t Time) Render() (int, string) {
	return t.Order - 1, t.Block.Render()
}

func (p Prompt) Render() (int, string) {
	if ExitCode == "0" {
		p.Block.Formatting = `\[\033[0;32m\]`
	} else {
		p.Block.Formatting = `\[\033[0;31m\]`
	}

	return p.Order - 1, p.Block.Render()
}

// KubeConfig represents a kubectl config file at ~/.kube/config
type KubeConfig struct {
	Contexts []struct {
		Context struct {
			Namespace string `yaml:"namespace"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
}

func (c Context) Render() (int, string) {
	var conf KubeConfig
	awsProfile := os.Getenv("AWS_PROFILE")
	kubeConfigFile := os.Getenv("HOME") + "/.kube/config"
	nameSpace := ""

	bytes, err := ioutil.ReadFile(kubeConfigFile)
	if err != nil {
		return c.Order - 1, ""
	}

	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return c.Order - 1, ""
	}

	if conf.CurrentContext != "" {
		for _, context := range conf.Contexts {
			if context.Name == conf.CurrentContext {
				nameSpace = context.Context.Namespace
				break
			}

		}

		switch {
		case nameSpace != "" && conf.CurrentContext == awsProfile:
			c.Block.Content = conf.CurrentContext + "|" + nameSpace
		case conf.CurrentContext == awsProfile && nameSpace == "":
			c.Block.Content = conf.CurrentContext
		case conf.CurrentContext != awsProfile:
			c.Block.Content = "[a]" + awsProfile + "[k]" + conf.CurrentContext
		}
	}

	return c.Order - 1, c.Block.Render()
}

func (g Git) Render() (int, string) {
	path := strings.Split(Dir, "/")
	targetDir := ""
	parentGit := false
	for i, _ := range path {
		if i == 0 {
			targetDir = "/"
		} else {
			targetDir = strings.Join(path[0:i+1], "/")
		}
		if gitCheck(targetDir) {
			parentGit = true
			break
		}
	}

	if parentGit {
		repo, err := git.PlainOpen(targetDir)
		if err != nil {
			return 0, ""
		}

		head, err := repo.Head()
		if err == nil {
			headStr := head.Name()
			branch := strings.Replace(string(headStr), "refs/heads/", "", 1)

			wt, _ := repo.Worktree()
			status, _ := wt.Status()

			if status.IsClean() {
				g.Block.Formatting = `\[\033[0;32m\]`
			} else {
				g.Block.Formatting = `\[\033[0;31m\]`
			}
			g.Block.Content = branch
		} else {
			g.Block.Content = "empty"
		}
	} else {
		return 0, ""
	}

	return g.Order - 1, g.Block.Render()
}

func gitCheck(dir string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if file.Name() == ".git" {
			return true
		}
	}
	return false
}

func (u User) Render() (int, string) {
	if u.Host {
		u.Block.Content = `\u@\h`
	} else {
		u.Block.Content = `\u`
	}
	return u.Order - 1, u.Block.Render()
}

func (b *Block) Render() string {
	var rendered strings.Builder
	rendered.WriteString(b.Formatting)
	rendered.WriteString(b.Prefix)
	rendered.WriteString(b.Content)
	rendered.WriteString(b.Suffix)
	return rendered.String()
}

func (p Pwd) Render() (int, string) {
	home := os.Getenv("HOME")
	if p.MaxDepth == 0 {
		if value, ok := os.LookupEnv("GOPS1_PWD_DEPTH"); ok {
			p.MaxDepth, _ = strconv.Atoi(value)
		}
	}

	modPwd := strings.Replace(Dir, home, "~", 1)
	path := strings.Split(modPwd, "/")

	if len(path) > p.MaxDepth {
		if len(path)-p.MaxDepth > 1 {
			path = append(path[:1], path[len(path)-p.MaxDepth+1:]...)
			path[0] = path[0] + "/..."
		} else {
			path[1] = "..."
		}

		p.Block.Content = strings.Join(path, "/")
	} else {
		p.Block.Content = modPwd
	}

	return p.Order - 1, p.Block.Render()
}
