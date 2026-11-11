package Lyger

const TEST_MAIN_BYTES = `$$HEAD<./test/main.l:1>HEAD$$Use Catalog's Print from ` + "`catalog`" + ` as "Print".
$$HEAD<./test/main.l:2>HEAD$$"x" is ` + "`Hello World!`" + `.
$$HEAD<./test/main.l:4>HEAD$$Create "HelloWorldApp" as a New Basic App.
$$HEAD<./test/main.l:5>HEAD$$The App Prints x.`

/*
// SCANNER TEST HELPERS
*/

func getWantedImportStatements(testType string) []*importStatement {
	var stmts []*importStatement
	if testType == "basic" {
		stmt := importStatement{
			statementBytes: []byte("$$HEAD<./test/main.l:1>HEAD$$Use Catalog's Print from `catalog` as \"Print\".\n"),
			module:         "Catalog",
			keyWord:        "Print",
			filePath:       "catalog",
			alias:          "Print",
		}
		stmts = append(stmts, &stmt)
	}
	return stmts
}
