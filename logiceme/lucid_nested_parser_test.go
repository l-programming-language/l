package Logiceme

import (
	"reflect"
	"strings"
	"testing"
)

func nestedParserConditional(body string) string {
	return "(/={tags: Meta.Tags}=={if Meta.Key != Void}=={" + body + "}=/)"
}

func TestLucidNestedParserPreservesCompleteBody(t *testing.T) {
	child := nestedParserConditional("inside")
	for _, tc := range []struct {
		name, body string
		children   []string
	}{
		{"mixed literal spans", "before" + child + "between" + child + "after", []string{child, child}},
		{"literal edge delimiters", "{{=leading" + child + "trailing=}}", []string{child}},
		{"CRLF and Unicode", "\r\n雪\t" + child + "\r\n光\r\n", []string{child}},
		{"empty", "", nil},
		{"literal braces", "{{body}}", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := nestedParserConditional(tc.body)
			tags := RecursiveSubTagParser(source, &TagMeta{})
			if len(tags) != 1 || len(tags[0].Expressions) != 3 {
				t.Fatalf("expected one tag with three expressions: %#v", tags)
			}
			if got := tags[0].Expressions[2].ExpressionString; got != tc.body {
				t.Errorf("body = %q; want original bytes %q", got, tc.body)
			}
			var children []string
			for _, child := range tags[0].Expressions[2].ExpressionSubTags {
				children = append(children, child.ReplaceString)
			}
			if !reflect.DeepEqual(children, tc.children) {
				t.Errorf("children = %q; want %q", children, tc.children)
			}
			wantExpressions := []string{"={tags: Meta.Tags}=", "={if Meta.Key != Void}=", "={" + tc.body + "}="}
			if got := lucidParser(strings.TrimSuffix(strings.TrimPrefix(source, "(/"), "/)"), "SubExpression"); !reflect.DeepEqual(got, wantExpressions) {
				t.Errorf("framed expressions = %q; want %q", got, wantExpressions)
			}
		})
	}
}

func TestLucidNestedParserPreservesEveryDepth(t *testing.T) {
	leaf := nestedParserConditional("leaf")
	middleBody := "middle-before" + leaf + "middle-after"
	middle := nestedParserConditional(middleBody)
	outerBody := "outer-before" + middle + "outer-after"
	tags := RecursiveSubTagParser(nestedParserConditional(outerBody), &TagMeta{})
	for depth, want := range []string{outerBody, middleBody, "leaf"} {
		if len(tags) != 1 || len(tags[0].Expressions) != 3 {
			t.Fatalf("depth %d: expected one tag with three expressions", depth)
		}
		if got := tags[0].Expressions[2].ExpressionString; got != want {
			t.Errorf("depth %d body = %q; want %q", depth, got, want)
		}
		tags = tags[0].Expressions[2].ExpressionSubTags
	}
	if len(tags) != 0 {
		t.Errorf("leaf has unexpected children: %#v", tags)
	}
}

func TestLucidNestedParserPreservesMainBodyEdges(t *testing.T) {
	for _, body := range []string{"{{body}}", "-leading trailing-", "{}", ""} {
		tags := CompileMainTags(lucidFramingTemplate(body))
		if len(tags) != 1 || len(tags[0].Expressions) != 3 {
			t.Fatalf("body %q: expected one main tag with three expressions", body)
		}
		if got := tags[0].Expressions[2].ExpressionString; got != body {
			t.Errorf("body = %q; want original bytes %q", got, body)
		}
	}
}

func TestLucidNestedParserRefreshesConjoinmentChildren(t *testing.T) {
	child := nestedParserConditional("[|Meta.Calculation|]")
	body := "before" + child + "after"
	tags := CompileMainTags(lucidFramingTemplate(body))
	if len(tags) != 1 || len(tags[0].Expressions) != 3 {
		t.Fatal("expected one main tag with three expressions")
	}
	meta := MetaDict{}
	meta.SetAdd("Calculation", "", "Conjoinment")
	tag := injectMetaConjoinments(tags[0], meta)
	wantChild := strings.ReplaceAll(child, "[|Meta.Calculation|]", "[|Conjoinments.Calculation|]")
	if got := tag.Expressions[2].ExpressionString; got != "before"+wantChild+"after" {
		t.Errorf("rewritten body = %q", got)
	}
	children := tag.Expressions[2].ExpressionSubTags
	if len(children) != 1 {
		t.Fatalf("rewritten body has %d children; want one", len(children))
	}
	if children[0].ReplaceString != wantChild || children[0].Expressions[2].ExpressionString != "[|Conjoinments.Calculation|]" {
		t.Errorf("child tree was not refreshed after metadata rewrite: %#v", children[0])
	}
}
