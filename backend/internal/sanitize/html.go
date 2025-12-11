package sanitize

import (
	"regexp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var (
	policy     *bluemonday.Policy
	policyOnce sync.Once
)

// getPolicy returns a singleton bluemonday policy configured for TipTap editor output
func getPolicy() *bluemonday.Policy {
	policyOnce.Do(func() {
		policy = bluemonday.NewPolicy()

		// Allow basic formatting tags
		policy.AllowElements("p", "br", "strong", "em", "u", "s")

		// Allow headings
		policy.AllowElements("h1", "h2", "h3")

		// Allow lists
		policy.AllowElements("ul", "ol", "li")

		// Allow blockquote
		policy.AllowElements("blockquote")

		// Allow code blocks with language attribute
		policy.AllowElements("code")
		policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-[\w-]+$`)).OnElements("code")
		policy.AllowElements("pre")
		policy.AllowAttrs("data-language").OnElements("pre")

		// Allow spans with specific classes (for syntax highlighting)
		policy.AllowElements("span")
		policy.AllowAttrs("class").Matching(regexp.MustCompile(`^hljs[\w-]*$`)).OnElements("span")

		// Allow links with safe protocols only
		policy.AllowElements("a")
		policy.AllowAttrs("href").OnElements("a")
		policy.AllowAttrs("target").Matching(regexp.MustCompile(`^_blank$`)).OnElements("a")
		policy.AllowAttrs("rel").Matching(regexp.MustCompile(`^(noopener|noreferrer|nofollow)(\s+(noopener|noreferrer|nofollow))*$`)).OnElements("a")
		policy.AllowAttrs("class").OnElements("a")
		policy.AllowURLSchemes("http", "https", "mailto")
		policy.RequireNoFollowOnLinks(true)
		policy.RequireNoReferrerOnLinks(true)
	})

	return policy
}

// HTML sanitizes HTML content to prevent XSS attacks.
// It allows only the HTML tags and attributes that are expected from the TipTap editor.
func HTML(html string) string {
	if html == "" {
		return ""
	}
	return getPolicy().Sanitize(html)
}

// HTMLPtr sanitizes HTML content and returns a pointer, useful for optional fields.
func HTMLPtr(html *string) *string {
	if html == nil {
		return nil
	}
	sanitized := HTML(*html)
	return &sanitized
}
