### Overview

A simple site crawler.

### Installation

- `go get github.com/jamiealquiza/site-fetch`
- `go install github.com/jamiealquiza/site-fetch`

Binary will be found at $GOPATH/bin/site-fetch.

### Usage

<pre>
Usage of ./site-fetch:
  -depth=1: Recursive crawl depth
  -dump=false: Dump map to file
  -url="": Target URL
</pre>
 
Example:

<pre>
% ./site-fetch -url="http://google.com" -depth=1
{
  "http://google.com": {
    "assets": [
      "http://google.com/images/srpr/logo9w.png",
      "http://google.com/images/icons/product/chrome-48.png"
    ],
    "links": [
      "http://google.com/chrome/index.html",
      "http://google.com/language_tools",
      "http://google.com/intl/en/about.html",
      "http://google.com/intl/en/policies/privacy",
      "http://google.com/preferences",
      "http://google.com/advanced_search",
      "http://google.com/intl/en/ads",
      "http://google.com/services",
      "http://google.com/intl/en/policies/terms"
    ]
  }
}
</pre>
