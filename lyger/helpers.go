package Lyger

const TEST_MAIN_BYTES = `$$HEAD<./test/main.l:1>HEAD$$Use IO's Print from ` + "`io`" + ` as "Print".
$$HEAD<./test/main.l:2>HEAD$$"x" is ` + "`Hello World!`" + `.
$$HEAD<./test/main.l:4>HEAD$$Create "HelloWorldApp" as a New Basic App.
$$HEAD<./test/main.l:5>HEAD$$The App Prints x.
$$HEAD<./test/main.l:6>HEAD$$Execute it.
$$HEAD<./test/main.l:7>HEAD$$If it has an Error, exit with the Error, else define "y" as it's Output.
$$HEAD<./test/main.l:8>HEAD$$If x isnt y, exit with an error of ` + "`text values of both x & y should be Hello World`" + `.
$$HEAD<./test/main.l:9>HEAD$$Otherwise, show y and then exit.`

/*
// SCANNER TEST HELPERS
*/

func getWantedImportStatements(testType string) []*importStatement {
	var stmts []*importStatement
	if testType == "basic" {
		stmt := importStatement{
			statementBytes: []byte("$$HEAD<./test/main.l:1>HEAD$$Use IO's Print from `io` as \"Print\".\n"),
			module:         "IO",
			keyWord:        "Print",
			filePath:       "io",
			alias:          "Print",
		}
		stmts = append(stmts, &stmt)
	}
	return stmts
}
