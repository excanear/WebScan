package templates

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Template is a YAML-based detection template.
type Template struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Severity    string    `yaml:"severity"` // info, low, medium, high, critical
	Description string    `yaml:"description"`
	Reference   []string  `yaml:"reference"`
	Tags        []string  `yaml:"tags"`
	Requests    []Request `yaml:"requests"`
}

// Request defines an HTTP probe.
type Request struct {
	Method           string            `yaml:"method"`
	Path             string            `yaml:"path"`
	Headers          map[string]string `yaml:"headers"`
	Body             string            `yaml:"body"`
	FollowRedirect   bool              `yaml:"follow_redirects"`
	Matchers         []Matcher         `yaml:"matchers"`
	MatcherCondition string            `yaml:"matcher-condition"` // and | or (default or)
}

// Matcher specifies how to detect a finding.
type Matcher struct {
	Type      string   `yaml:"type"` // status | word | regex | header
	Status    []int    `yaml:"status"`
	Words     []string `yaml:"words"`
	Regex     []string `yaml:"regex"`
	Part      string   `yaml:"part"`      // body | header | all (default body)
	Condition string   `yaml:"condition"` // and | or (default or)
}

// MatchResult holds a template match finding.
type MatchResult struct {
	TemplateID  string   `json:"template_id"`
	Name        string   `json:"name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags,omitempty"`
}

// LoadDir loads all *.yaml templates from dir (recursive).
func LoadDir(dir string) ([]Template, error) {
	var out []Template
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		t, err := LoadFile(path)
		if err != nil {
			return fmt.Errorf("template %s: %w", path, err)
		}
		out = append(out, t)
		return nil
	})
	return out, err
}

// LoadFile loads a single template file.
func LoadFile(path string) (Template, error) {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return Template{}, err
	}
	var t Template
	if err := yaml.Unmarshal(data, &t); err != nil {
		return Template{}, err
	}
	return t, nil
}

// Run executes a template against baseURL and returns any matches.
func Run(t Template, baseURL string, timeout time.Duration) []MatchResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var results []MatchResult

	for _, req := range t.Requests {
		matched, url := runRequest(t, req, baseURL, timeout)
		if matched {
			results = append(results, MatchResult{
				TemplateID:  t.ID,
				Name:        t.Name,
				Severity:    t.Severity,
				Description: t.Description,
				URL:         url,
				Tags:        t.Tags,
			})
		}
	}
	return results
}

// RunAll executes all templates against baseURL.
func RunAll(templates []Template, baseURL string, timeout time.Duration) []MatchResult {
	var all []MatchResult
	for _, t := range templates {
		all = append(all, Run(t, baseURL, timeout)...)
	}
	return all
}

func runRequest(t Template, req Request, baseURL string, timeout time.Duration) (bool, string) {
	method := strings.ToUpper(req.Method)
	if method == "" {
		method = "GET"
	}
	targetURL := strings.TrimRight(baseURL, "/") + req.Path

	client := &http.Client{Timeout: timeout}
	if !req.FollowRedirect {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		return false, ""
	}
	httpReq.Header.Set("User-Agent", "WebScan/1.0")
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	bodyStr := string(bodyBytes)

	matched := evalMatchers(req.Matchers, req.MatcherCondition, resp, bodyStr)
	return matched, targetURL
}

func evalMatchers(matchers []Matcher, condition string, resp *http.Response, body string) bool {
	if len(matchers) == 0 {
		return false
	}
	topAnd := strings.ToLower(condition) == "and"
	for _, m := range matchers {
		result := evalMatcher(m, resp, body)
		if topAnd && !result {
			return false
		}
		if !topAnd && result {
			return true
		}
	}
	return topAnd // all matched for "and", none matched for "or"
}

func evalMatcher(m Matcher, resp *http.Response, body string) bool {
	switch strings.ToLower(m.Type) {
	case "status":
		for _, s := range m.Status {
			if resp.StatusCode == s {
				return true
			}
		}
		return false

	case "word":
		target := matchPart(m.Part, resp, body)
		useAnd := strings.ToLower(m.Condition) == "and"
		for _, w := range m.Words {
			found := strings.Contains(target, w)
			if useAnd && !found {
				return false
			}
			if !useAnd && found {
				return true
			}
		}
		return useAnd

	case "regex":
		target := matchPart(m.Part, resp, body)
		useAnd := strings.ToLower(m.Condition) == "and"
		for _, pattern := range m.Regex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			found := re.MatchString(target)
			if useAnd && !found {
				return false
			}
			if !useAnd && found {
				return true
			}
		}
		return useAnd

	case "header":
		// match against a specific header value
		for _, w := range m.Words {
			for _, vals := range resp.Header {
				for _, v := range vals {
					if strings.Contains(strings.ToLower(v), strings.ToLower(w)) {
						return true
					}
				}
			}
		}
		return false
	}
	return false
}

func matchPart(part string, resp *http.Response, body string) string {
	switch strings.ToLower(part) {
	case "header", "headers":
		var sb strings.Builder
		for k, vals := range resp.Header {
			for _, v := range vals {
				sb.WriteString(k + ": " + v + "\n")
			}
		}
		return sb.String()
	case "all":
		var sb strings.Builder
		for k, vals := range resp.Header {
			for _, v := range vals {
				sb.WriteString(k + ": " + v + "\n")
			}
		}
		sb.WriteString(body)
		return sb.String()
	default:
		return body
	}
}
