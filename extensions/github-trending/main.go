package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("GitHub Trending Extension server starting on 0.0.0.0:8081")
	// Listen on all interfaces within the container
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Set the required header for Glance
	w.Header().Set("Widget-Content-Type", "html")
	// Add headers to prevent browser caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	// Fetch the trending page
	res, err := http.Get("https://github.com/trending")
	if err != nil {
		log.Printf("Error fetching github trending: %v", err)
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Error fetching github trending: status code %d", res.StatusCode)
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		http.Error(w, "Error parsing data", http.StatusInternalServerError)
		return
	}

	// Start building the HTML output
	var htmlBuilder strings.Builder

	// Add CSS for styling list items
	htmlBuilder.WriteString(`
<style>
  .gh-trending-list .list-item {
    border-bottom: 2px solid var(--color-border); /* Solid bottom border */
    padding: 10px; /* Add padding inside the box */
    margin-bottom: 10px; /* Adjust margin for spacing between items */
    background-color: rgba(255, 255, 255, 0.05); /* Subtle background color */
  }
  .gh-trending-list .list-item:last-child {
    margin-bottom: 0; /* Remove margin from the last item */
    border-bottom: none; /* Remove border from the last item */
    padding: 10px; /* Add padding to the last item */
  }
  .gh-trending-list .repo-language-color {
	display: inline-block;
	width: 10px;
	height: 10px;
	border-radius: 50%;
	margin-right: 4px;
	vertical-align: middle;
	/* Default color, ideally replaced by actual language color */
	background-color: var(--color-text-secondary);
  }
</style>
`)

	// Add a class to the main list for scoping the styles
	htmlBuilder.WriteString(`<ul class="list gh-trending-list">`) // Removed list-gap-15 as margin-bottom handles spacing

	// Find the repository items
	doc.Find("article.Box-row").Each(func(i int, s *goquery.Selection) {
		// Extract data for each repository
		repoLink := s.Find("h2 a")
		repoName := strings.TrimSpace(repoLink.Text())
		repoURL, _ := repoLink.Attr("href")
		description := strings.TrimSpace(s.Find("p.col-9").Text())
		starsToday := strings.TrimSpace(s.Find("span.d-inline-block.float-sm-right").Text())
		language := strings.TrimSpace(s.Find("span[itemprop='programmingLanguage']").Text())
		totalStars := strings.TrimSpace(s.Find("a[href$='/stargazers']").Text())
		forks := strings.TrimSpace(s.Find("a[href$='/forks']").Text())

		if repoName == "" || repoURL == "" {
			return // Skip if essential info is missing
		}

		// Build HTML list item for the repository
		htmlBuilder.WriteString(`<li class="list-item">`)
		htmlBuilder.WriteString(fmt.Sprintf(`<a class="size-h4 color-highlight block text-truncate" href="https://github.com%s" target="_blank">%s</a>`, repoURL, repoName))

		if description != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`<p class="color-paragraph size-h5 margin-top-5">%s</p>`, description))
		}

		htmlBuilder.WriteString(`<ul class="list-horizontal-text size-h6 margin-top-10">`)
		if language != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`<li><span class="repo-language-color"></span> %s</li>`, language))
		}
		if totalStars != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`<li>⭐ %s</li>`, totalStars))
		}
		if forks != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`<li><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="1em" height="1em" fill="currentColor"><path d="M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.5.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-5 8.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z"></path></svg> %s</li>`, forks))
		}
		if starsToday != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`<li>⭐ %s</li>`, starsToday))
		}
		htmlBuilder.WriteString(`</ul>`) // End horizontal list

		htmlBuilder.WriteString(`</li>`) // End list item
	})

	htmlBuilder.WriteString(`</ul>`) // End main list

	// Send the generated HTML
	fmt.Fprint(w, htmlBuilder.String())
}
