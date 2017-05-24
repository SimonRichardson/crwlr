package document_test

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/SimonRichardson/crwlr/pkg/document"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/html"
)

func ExampleDocument_Walk_Links() {
	body := `
<!DOCTYPE html>
<html>
<head>
<title>Title</title>
</head>
<body>
<h1>Heading</h1>
<a href="/link">link</a>
<a href="http://url.com/link1">link1</a>
<a href="/link2">link2</a>
<a href="http://google.com/link3">link3</a>
</body>
`

	u, err := url.Parse("http://url.com")
	if err != nil {
		panic(err)
	}

	node, err := html.Parse(strings.NewReader(body))
	if err != nil {
		panic(err)
	}

	doc := document.NewDocument(u, node, log.NewNopLogger())
	doc.WalkLinks(func(url *url.URL) error {
		fmt.Printf("The url is: %s\n", url.String())
		return nil
	})

	// Output:
	// The url is: http://url.com/link
	// The url is: http://url.com/link1
	// The url is: http://url.com/link2
	// The url is: http://google.com/link3
}
