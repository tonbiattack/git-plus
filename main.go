package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tonbiattack/git-plus/cmd"
)

func main() {
	// 実行ファイル名からサブコマンドを推測
	execName := filepath.Base(os.Args[0])

	// "git-" で始まる場合、それをサブコマンドとして扱う
	if strings.HasPrefix(execName, "git-") {
		// "git-newbranch" → "newbranch"
		subCommand := strings.TrimPrefix(execName, "git-")

		// os.Argsを書き換えて、サブコマンドを挿入
		// 例: ["git-newbranch", "arg1"] → ["git-plus", "newbranch", "arg1"]
		newArgs := make([]string, len(os.Args)+1)
		newArgs[0] = os.Args[0]
		newArgs[1] = subCommand
		copy(newArgs[2:], os.Args[1:])
		os.Args = newArgs
	}

	cmd.Execute()
}
