package browser

import (
	"context"
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
)

type mockPlaywrightRunner struct {
	err error
	pw  *mockPlaywright
}

func (m *mockPlaywrightRunner) Run() (playwrightImpl, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pw, nil
}

type mockPlaywright struct {
	stopErr error
	bt      *mockBrowserType
}

func (m *mockPlaywright) Stop() error { return m.stopErr }
func (m *mockPlaywright) Chromium() playwrightBrowserType {
	return m.bt
}

type mockBrowserType struct {
	err error
	b   *mockBrowser
}

func (m *mockBrowserType) Launch(options ...playwright.BrowserTypeLaunchOptions) (playwrightBrowser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.b, nil
}

type mockBrowser struct {
	closeErr error
	pageErr  error
	p        *mockPage
}

func (m *mockBrowser) Close() error { return m.closeErr }
func (m *mockBrowser) NewPage(options ...playwright.BrowserNewPageOptions) (playwrightPage, error) {
	if m.pageErr != nil {
		return nil, m.pageErr
	}
	return m.p, nil
}

type mockPage struct {
	gotoErr error
	l       *mockLocator
}

func (m *mockPage) Goto(url string, options ...playwright.PageGotoOptions) (playwright.Response, error) {
	if m.gotoErr != nil {
		return nil, m.gotoErr
	}
	return nil, nil
}
func (m *mockPage) Locator(selector string, options ...playwright.PageLocatorOptions) playwrightLocator {
	return m.l
}

type mockLocator struct {
	text string
	err  error
}

func (m *mockLocator) TextContent(options ...playwright.LocatorTextContentOptions) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.text, nil
}

func TestFetchText_Success(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			bt: &mockBrowserType{
				b: &mockBrowser{
					p: &mockPage{
						l: &mockLocator{
							text: "mocked content",
						},
					},
				},
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	content, err := f.FetchText(context.Background(), "http://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "mocked content", content)
}

func TestFetchText_PlaywrightRunError(t *testing.T) {
	runner := &mockPlaywrightRunner{
		err: fmt.Errorf("run error"),
	}

	f := &playwrightFetcher{runner: runner}
	_, err := f.FetchText(context.Background(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not start playwright")
}

func TestFetchText_LaunchError(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			bt: &mockBrowserType{
				err: fmt.Errorf("launch error"),
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	_, err := f.FetchText(context.Background(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not launch browser")
}

func TestFetchText_NewPageError(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			bt: &mockBrowserType{
				b: &mockBrowser{
					pageErr: fmt.Errorf("new page error"),
				},
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	_, err := f.FetchText(context.Background(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not create page")
}

func TestFetchText_GotoError(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			bt: &mockBrowserType{
				b: &mockBrowser{
					p: &mockPage{
						gotoErr: fmt.Errorf("goto error"),
					},
				},
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	_, err := f.FetchText(context.Background(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not goto")
}

func TestFetchText_TextContentError(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			bt: &mockBrowserType{
				b: &mockBrowser{
					p: &mockPage{
						l: &mockLocator{
							err: fmt.Errorf("text content error"),
						},
					},
				},
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	_, err := f.FetchText(context.Background(), "http://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not extract text content")
}

func TestFetchText_StopAndCloseErrors(t *testing.T) {
	runner := &mockPlaywrightRunner{
		pw: &mockPlaywright{
			stopErr: fmt.Errorf("stop error"),
			bt: &mockBrowserType{
				b: &mockBrowser{
					closeErr: fmt.Errorf("close error"),
					p: &mockPage{
						l: &mockLocator{
							text: "content",
						},
					},
				},
			},
		},
	}

	f := &playwrightFetcher{runner: runner}
	content, err := f.FetchText(context.Background(), "http://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "content", content)
}
