package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
	//ps1UserUost := `\[\e[1;37m\][\u@\h]`
	ps1User := `\[\e[1;37m\][\u]`
	ps1NewLine := `\n`
	ps1Time := `\[\e[1;32m\][\A]`
	//ps1Dollar := `\[\033[0;32m\]$`

	pwd, _ := os.Getwd()
	ps1Ctx, _ := getContext()
	PS1 := ps1TopBracket + ps1User + ps1Line + getPwd(pwd) + gitInfo(pwd) + ps1NewLine + ps1BottomBracket + ps1Ctx + ps1Time + dollarPrompt(exitCode) + ps1Space

	return PS1, nil
}

func getPwd(pwd string) string {
	ps1Format := `\[\e[0;93m\](`
	home := os.Getenv("HOME")
	modPwd := strings.Replace(pwd, home, "~", 1)
	return ps1Format + modPwd + ")"
}

func gitInfo(pwd string) string {
	ps1Space := `\[\033[00m\] `
	gitPrompt := `\[\033[0;32m\](`
	repo, err := git.PlainOpen(pwd)
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
		case nameSpace != "" && conf.CurrentContext == awsProfile:
			ctx = conf.CurrentContext + "|" + nameSpace
		case conf.CurrentContext == awsProfile && nameSpace == "":
			ctx = conf.CurrentContext
		case conf.CurrentContext != awsProfile:
			ctx = "[a]" + awsProfile + "[k]" + conf.CurrentContext
		}
	}

	ctxString := ctxFormatting + ctx + ")"

	return ctxString, nil
}
