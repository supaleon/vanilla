package tests

import (
	//"github.com/supaleon/vanilla/internal/lang/html2"
	"github.com/tdewolff/parse/v2"
	html2 "github.com/tdewolff/parse/v2/html"
	"io"
	"strings"
	"testing"
)

const testcase = `aaa<Hello id='text' aa ><o:AllowPNG/></Hello  > </div </div

`

func TestPrint2(t *testing.T) {
	tokenizer := html2.NewLexer(parse.NewInput(strings.NewReader(testcase)))
	for {
		tt, data := tokenizer.Next()
		switch tt {
		case html2.ErrorToken:
			println(tokenizer.Err().Error())
			if err := tokenizer.Err(); err != nil && err != io.EOF {
				t.Fatal(tokenizer.Err())
			}
			return
		case html2.CommentToken:
			println("CommentToken:", string(data))
		case html2.DoctypeToken:
			println("DoctypeToken:", string(data))
		case html2.StartTagToken:
			//tn, _ := tokenizer.TagName()
			println("StartTagToken:", string(data))
		case html2.EndTagToken:
			//tn, _ := tokenizer.TagName()
			println("EndTagToken:", string(data))
		case html2.TextToken:
			println("TextToken:", string(tokenizer.Text()))
		case html2.StartTagVoidToken:
			println("SelfClosingTagToken:", string(data))
		case html2.StartTagCloseToken:
			println("StartTagCloseToken:", string(data))
		case html2.SvgToken:
			println("SvgToken:", string(data))
		case html2.AttributeToken:
			println("AttributeToken:", string(data))
		default:
			println("11111")
			return
		}
	}
}

//func TestPrint(t *testing.T) {
//	tokenizer := html2.NewTokenizer(strings.NewReader(testcase))
//	//tokenizer.AllowCDATA(false)
//	for {
//		tt := tokenizer.Next()
//		switch tt {
//		case html2.ErrorToken:
//			println(tokenizer.Err().Error())
//			if err := tokenizer.Err(); err != nil && err != io.EOF {
//				t.Fatal(tokenizer.Err())
//			}
//			return
//		case html2.CommentToken:
//			println("CommentToken:", tokenizer.Token().Data)
//		case html2.DoctypeToken:
//			println("DoctypeToken:", string(tokenizer.Text()))
//		case html2.StartTagToken:
//			tn, _ := tokenizer.TagName()
//			println("StartTagToken:", string(tn), ":atom:", tokenizer.Token().DataAtom.String())
//		case html2.EndTagToken:
//			tn, _ := tokenizer.TagName()
//			println("EndTagToken:", string(tn), "raw:", string(tokenizer.Raw()), ":atom:", tokenizer.Token().DataAtom.String())
//		case html2.TextToken:
//			println("TextToken:", string(tokenizer.Text()), "textRaw:", string(tokenizer.Raw()))
//		case html2.SelfClosingTagToken:
//			tn, _ := tokenizer.TagName()
//			println("SelfClosingTagToken:", string(tn), ":atom:", tokenizer.Token().DataAtom.String())
//		default:
//			println("11111")
//			return
//		}
//	}
//}

var invalidTags = []string{
	"style",
	"script",
} /**/
