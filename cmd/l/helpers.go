package main

import (
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
	"os"
	"path/filepath"
)

const INITTEST = `Use IO's Print from ` + "`io`" + ` as "Print".
"x" is ` + "`Hello World!`" + `.
*** This is a comment ***
Create "HelloWorldApp" as a New Basic App.
The App Prints x.                                     *** This prints ` + "`Hello World`" + ` once HelloWorldApp is go built and ran ***
Execute it.                                            *** 1) Builds & Runs the HelloWorldApp's exe file and then 2) runs it ***
If it has an Error, exit with the Error, else define "y" as it's Output.
If x isnt y, exit with an error of ` + "`text values of both x & y should be Hello World`" + `.
Otherwise, show y and then exit.`

// COMPTEST is the exact originating-user-source projection, including comments.
// Builtin and dictionary bodies have resource/identity tests, not copied goldens.
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
