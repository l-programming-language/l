package main

import (
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
	"os"
	"path/filepath"
)

const INITTEST = `"x" is ` + "`Hello World!`" + `.
Create "HelloWorldApp" as a New Basic App.
The App Prints x.`

// COMPTEST is the exact originating-user-source projection.
const COMPTEST = INITTEST

func getTestOptions() *Lyger.Options {
	root, err := filepath.Abs("../..")
	if err != nil {
		return nil
	}
	return &Lyger.Options{Language: Lyger.DefaultLanguage, AssetRoot: root, Resources: os.DirFS(root)}
}

func getTestLogService(configSettings *Lyger.Options) *Logiceme.LogicemeService {
	return Logiceme.NewLogicemeService(configSettings.Resources)
}

func getTestLygerService(configSettings *Lyger.Options) *Lyger.LygerService {
	return Lyger.NewLygerService(*configSettings)
}

// getTestCompiler
func getTestCompiler() *Lyger.Compilation {
	return &Lyger.Compilation{}
}
