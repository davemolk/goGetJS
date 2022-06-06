# goGetJS
## Cool Features
* Scrapes page for script tags, saving the src to a text file. It then visits each linked script and writes the contents to an individual file.
* If a script tag doesn't have an src attribute, go Get JS scrapes everything inbetween the script tags and saves it to a file.
* Launch a headless browser (playwright) to handle JS-heavy sites.
* Add an extra wait period to capture those longer loading async scripts.
* Scan each script for a specific term, a regular expression, or a list of terms (input as a text file).
* Leverages goroutines for fast processing.