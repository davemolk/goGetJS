# goGetJS

## Cool Features
* Scrapes page for script tags, saving the src to a text file. It then visits each src and writes the contents to an individual file.
* If a script tag doesn't have an src attribute, goGetJS scrapes everything between the script tags and saves it to a file.
* Launch a headless browser (playwright) to handle JS-heavy sites.
* Add an extra wait period to capture those longer loading async scripts.
* Scan each script for a specific term, a regular expression, or a list of terms (input as a text file).
* Leverages goroutines for fast processing.

### Additional Notes
* Flags are pretty self-explanatory and include: url, timeout, browser, extraWait, term, regex, file.
* By default, playwright waits until the network has been idle for at least 500ms before considering the page loaded. You can use the -extraWait flag (measured in seconds) to extend this duration.
* When naming js files, goGetJS attempts to use the format <name.js>. In the event that it can't, it will grab a long chunk of [\w-].
* Include just one term per line when inputting a file of terms to search.