package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	PS1 := ""
	exitCode := "0"
	if len(os.Args) >= 2 {
		exitCode = os.Args[1]
	}
	if os.Getenv("GOPS1_MINIMAL") == "true" {
		PS1 = buildMinimalPS1(exitCode)
	} else {
		PS1, _ = buildPS1(exitCode)
	}

	fmt.Println(PS1)
}

func buildMinimalPS1(exitCode string) string {
	promptFormattingGreen := `\[\033[0;32m\]`
	promptFormattingRed := `\[\033[0;31m\]`
	ps1Space := `\[\033[00m\] `
	PS1 := ""
	if exitCode == "0" {
		PS1 = promptFormattingGreen + "~" + ps1Space
	} else {
		PS1 = promptFormattingRed + "~" + ps1Space
	}

	return PS1

}

func buildPS1(exitCode string) (string, error) {
	ps1TopBracket := `\[\e[0;36m\]┌─`
	ps1Line := `\[\e[0m\]\[\e[0;36m\]─`
	ps1BottomBracket := `\[\e[0;36m\]└─`
	ps1Space := `\[\033[00m\] `
	ps1NewLine := `\n`
	ps1Time := `\[\e[1;32m\][\A]`
	pwd, _ := os.Getwd()
	ps1Ctx, _ := getContext()

	var ps1 strings.Builder
	ps1.WriteString(ps1TopBracket)
	ps1.WriteString(ps1User())
	ps1.WriteString(ps1Line)
	ps1.WriteString(getPwd(pwd))
	ps1.WriteString(gitInfo(pwd))
	ps1.WriteString(ps1NewLine)
	ps1.WriteString(ps1BottomBracket)
	ps1.WriteString(ps1Ctx)
	ps1.WriteString(ps1Time)
	ps1.WriteString(dollarPrompt(exitCode))
	ps1.WriteString(ps1Space)

	return ps1.String(), nil
}

func ps1User() string {
	if os.Getenv("GOPS1_HOST") == "true" {
		return `\[\e[1;37m\][\u@\h]`
	}
	return `\[\e[1;37m\][\u]`
}

func getPwd(pwd string) string {
	ps1Format := `\[\e[0;93m\](`
	home := os.Getenv("HOME")
	out := ""

	pwdMaxDepth := 20
	if value, ok := os.LookupEnv("GOPS1_PWD_DEPTH"); ok {
		pwdMaxDepth, _ = strconv.Atoi(value)
	}

	modPwd := strings.Replace(pwd, home, "~", 1)
	path := strings.Split(modPwd, "/")

	if len(path) > pwdMaxDepth {
		if len(path)-pwdMaxDepth > 1 {
			path = append(path[:1], path[len(path)-pwdMaxDepth+1:]...)
			path[0] = path[0] + "/..."
		} else {
			path[1] = "..."
		}

		out = strings.Join(path, "/")
	} else {
		out = modPwd
	}
	return ps1Format + out + ")"
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

func gitInfo(pwd string) string {
	ps1Space := `\[\033[00m\] `
	gitPrompt := `\[\033[0;32m\](`
	targetDir := ""
	parentGit := false
	path := strings.Split(pwd, "/")

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
			return ""
		}

		head, err := repo.Head()
		if err == nil {
			headStr := head.Name()
			branch := strings.Replace(string(headStr), "refs/heads/", "", 1)

			wt, _ := repo.Worktree()
			status, _ := wt.Status()

			if status.IsClean() {
				gitPrompt = `\[\033[0;32m\](`
			} else {
				gitPrompt = `\[\033[0;31m\](`
			}

			return ps1Space + gitPrompt + branch + ")"
		} else {
			return ps1Space + `\[\033[0;31m\](` + "empty" + ")"
		}
	}
	return ""
}

func dollarPrompt(exitCode string) string {
	promptFormattingGreen := `\[\033[0;32m\]`
	promptFormattingRed := `\[\033[0;31m\]`
	if exitCode == "0" {
		return promptFormattingGreen + "$"
	} else {
		return promptFormattingRed + "$"
	}
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

func getContext() (string, error) {
	var conf KubeConfig
	ctxFormatting := `\[\e[0;36m\](`
	awsProfile := os.Getenv("AWS_PROFILE")
	kubeConfigFile := os.Getenv("HOME") + "/.kube/config"
	ctx := ""
	nameSpace := ""

	bytes, err := ioutil.ReadFile(kubeConfigFile)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return "", err
	}

	if conf.CurrentContext != "" {
		for _, context := range conf.Contexts {
			if context.Name == conf.CurrentContext {
				nameSpace = context.Context.Namespace
				break
			}

		}

		switch {
		case conf.CurrentContext == "rancher2":
			ctx = conf.CurrentContext
		case nameSpace != "" && conf.CurrentContext == awsProfile:
			ctx = conf.CurrentContext + "|" + nameSpace
		case nameSpace != "" && conf.CurrentContext == "rancher2":
			ctx = conf.CurrentContext
		case conf.CurrentContext == awsProfile && nameSpace == "":
			ctx = conf.CurrentContext
		case conf.CurrentContext != awsProfile:
			ctx = "[a]" + awsProfile + "[k]" + conf.CurrentContext
		}
	}

	ctxString := ctxFormatting + ctx + ")"

	return ctxString, nil
}
