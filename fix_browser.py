with open("server/pkg/tool/browser/browser.go", "r") as f:
    content = f.read()

content = content.replace("""type PageFetcher interface {
	FetchText(ctx context.Context, url string) (string, error)
}""", """type PageFetcher interface {
	// FetchText fetches the text content of a given URL.
	//
	// Summary: Retrieves the visible text content from a web page.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the request.
	//   - url (string): The URL of the web page to fetch.
	//
	// Returns:
	//   - string: The extracted text content from the page.
	//   - error: An error if the fetching or extraction fails.
	//
	// Errors:
	//   - Returns an error if the browser fails to start, navigate, or extract text.
	//
	// Side Effects:
	//   - May launch a browser process or make a network request.
	FetchText(ctx context.Context, url string) (string, error)
}""")

content = content.replace("""func (f *playwrightFetcher) FetchText(_ context.Context, url string) (string, error) {""", """// FetchText uses playwright to fetch the text content of a given URL.
//
// Summary: Retrieves the visible text content from a web page using playwright.
//
// Parameters:
//   - _ (context.Context): The context for the request (unused).
//   - url (string): The URL of the web page to fetch.
//
// Returns:
//   - string: The extracted text content from the page.
//   - error: An error if the fetching or extraction fails.
//
// Errors:
//   - Returns an error if playwright fails to start, browser fails to launch, navigation fails, or text extraction fails.
//
// Side Effects:
//   - Launches a headless chromium browser process and makes a network request.
func (f *playwrightFetcher) FetchText(_ context.Context, url string) (string, error) {""")

with open("server/pkg/tool/browser/browser.go", "w") as f:
    f.write(content)
