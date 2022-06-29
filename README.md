# goGetJS
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/davemolk/goGetJS)](https://goreportcard.com/report/github.com/davemolk/goGetJS)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/davemolk/goGetJS/issues)

goGetJS extracts, searches, and saves JavaScript files. Includes an optional browser (playwright) for dealing with JavaScript-heavy sites.
![demo](demo.gif)
## Overview
* goGetJS scrapes a given page for script tags, visits each src, and writes the contents to an individual file.
* All src are also saved to a text file.
* If a script tag doesn't include an src attribute, goGetJS scrapes everything between the script tags and writes the contents to an individual file.
* goGetJS (optionally) uses playwright to handle JavaScript-heavy sites and retrieve async scripts. Use -b=true.
* Add some extra waiting time with -ew to grab those longer loading async scripts.
* Use -w, -r, and -f, respectively, to scan each script for a specific word, with a regular expression, or with a list of words (input as a file).

## Example Usages
```
go run ./cmd/goGetJS -u=https://go.dev -b=true -f=search.txt
```
```
echo https://go.dev | goGetJS -b=true -f=search.txt
```

## Additional Notes
* Possible flags: -u (url), -t (timeout), -b (browser), -bt (browser timeout), -ew (extra wait), -w (word), -r (regex), -f (file).
* By default, playwright waits until the network has been idle for at least 500ms before considering the page loaded. Use the -ew flag (measured in seconds) for an additional wait period.
* When editing search.txt for your own use (or creating your own file), include just one term per line.
* goGetJS names JavaScript files with ```fName := regexp.MustCompile(`[\w-&]+(\.js)?$`)```. Most scripts play nice, but those that don't are still saved. Each saved script has the full URL prepended to the file.
* Occasionally, an src will link to an empty page. These are automatically retried and will sometimes get a script on that second attempt (which is searched and saved). More often, these are legitimately blank, causing the number of saved files printed to the terminal to be fewer than the number of processed files.

## Support
* Like goGetJS? Use it, star it, and share with your friends!
    - Let me know what you're up to so I can feature your work here.
* Want to see a particular feature? Found a bug? Question about usage or documentation?
    - Please raise an issue.
* Pull request?
    - Please discuss in an issue first. 

## Built With
* https://github.com/PuerkitoBio/goquery
* https://github.com/playwright-community/playwright-go

## License
* goGetJS is released under the MIT license. See [LICENSE](LICENSE) for details.