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
* goGetJS (optionally) uses playwright to handle JavaScript-heavy sites and async scripts. Use -browser=true.
* Use -extraWait to grab those longer loading async scripts.
* Use -term, -regex, and -file, respectively, to scan each script for a specific term, a regular expression, or a list of terms.

## Additional Notes
* Possible flags: url, timeout, browser, browserTimeout, extraWait, term, regex, file.
* By default, playwright waits until the network has been idle for at least 500ms before considering the page loaded. You can use the -extraWait flag (measured in seconds) for an additional wait period.
* When editing search.txt for your own use (or creating your own file), include just one term per line.
* goGetJS names JavaScript files with ```fName := regexp.MustCompile(`[\w-]+(\.js)?$`)```. Most scripts play nice, but those that don't are still saved.
* Occasionally, an src will link to an empty page. These are retried automatically (using a different client) and will sometimes show a script (which is then searched and saved). More often, these are actually blank, so you might find the number of saved scripts to be less than the number of processed scripts printed in the terminal. These blank pages tend to be metrics-related and generally unrelated to whatever investigations you're involved in with the JavaScript.

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