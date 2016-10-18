package main

import (
	"strings"
	"testing"
	//"fmt"
	"reflect"

	"golang.org/x/net/html"
)

var fakePage = strings.NewReader(`
		<html>
		<a href="http://someurl.com">link</a>
		<a href="http://someurl.com/about">link</a>
		<a href="http://anotherurl.com">link</a>
		<img src="http://someurl.com/someimg.jpg">
		</html>
		`)

var tokens = tokenize(fakePage)

func Test_tokenize(t *testing.T) {

	expected := []html.Attribute{
		html.Attribute{"", "href", "http://someurl.com"},
		html.Attribute{"", "href", "http://someurl.com/about"},
		html.Attribute{"", "href", "http://anotherurl.com"},
		html.Attribute{"", "src", "http://someurl.com/someimg.jpg"},
	}

	actual := tokens

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Got %s, expected %s", actual, expected)
	}

	/* // But that manual logic tho.
	if len(actual) != len(expected) {
		t.Fatalf("Got %s, expected %s", actual, expected)
	}

	// Short-circuit if slice isn't even the same length.
	same := true
	for i := range actual {
		if actual[i] != expected[i] {
			same = false
		}
	}


	if !same {
		t.Fatalf("Got %s, expected %s", actual, expected)
	}
	*/
}

func Test_mapAssetsAndLinks(t *testing.T) {
	expected := PageMap{
		"http://someurl.com",
		[]string{"http://someurl.com/someimg.jpg"},
		[]string{"http://someurl.com", "http://someurl.com/about"},
	}

	actual := mapAssetsAndLinks(tokens, "http://someurl.com")
	
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Got %s, expected %s", actual, expected)
	}
}