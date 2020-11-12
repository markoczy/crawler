# Crawler

A powerful Web Crawler based on Go and chromedp for experienced users.

## Features

- **Chromium based:** Renders and analyzes websites using chromium headless (using chromedp) to ensure that the pages are rendered just like in a web browser, this allows the crawler to analyze Javascript-Only pages just like normal html pages. Links are retreived by running JS scripts on the rendered page after the browser sends the "Dom Tree Loaded" event.
- **Recursive link scanning:** Visits a page and retreives all links from the page. Recursively visits all links up to the specified depth.
- **Recursive Download:** Downloads files from all retreived links.
- **Regex powered customizability:** Configure regular expressions to decide whick links to follow or download. Capture tokens from url naming patterns and bake them into your desired output file names.
- **HTTP Headers:** Add any http header by file or in the command line by the `-header` switch. Also supports easy basic auth with the `-auth` switch and easy user agent setting with the `-user-agent` switch.
- **URL Permutations:** URLs to scan can be configured by permutative scemes e.g. `myfile-[1-99]` would create an url for `myfile-1`, `myfile-2` ... `myfile-99`. Multiple permutative scemes in one url (such as `mypage-[a,b,c,d]/myfile-[1-99]`) are also supported.

