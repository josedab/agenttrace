package agenttrace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// PromptVersion represents a specific version of a prompt.
type PromptVersion struct {
	ID        string         `json:"id"`
	Version   int            `json:"version"`
	Prompt    string         `json:"prompt"`
	Config    map[string]any `json:"config"`
	Labels    []string       `json:"labels"`
	CreatedAt string         `json:"createdAt"`
}

// Compile compiles the prompt with variables.
func (p *PromptVersion) Compile(variables map[string]any) string {
	result := p.Prompt

	for key, value := range variables {
		// Support both {{var}} and {var} syntax
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), fmt.Sprint(value))
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), fmt.Sprint(value))
	}

	return result
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompileChat compiles the prompt as chat messages.
func (p *PromptVersion) CompileChat(variables map[string]any) []ChatMessage {
	compiled := p.Compile(variables)
	messages := []ChatMessage{}

	lines := strings.Split(strings.TrimSpace(compiled), "\n")
	roleRegex := regexp.MustCompile(`(?i)^(system|user|assistant|function):\s*(.*)$`)

	var currentRole string
	var currentContent []string

	for _, line := range lines {
		if matches := roleRegex.FindStringSubmatch(line); matches != nil {
			// Save previous message
			if currentRole != "" && len(currentContent) > 0 {
				messages = append(messages, ChatMessage{
					Role:    strings.ToLower(currentRole),
					Content: strings.TrimSpace(strings.Join(currentContent, "\n")),
				})
			}

			currentRole = matches[1]
			if matches[2] != "" {
				currentContent = []string{matches[2]}
			} else {
				currentContent = []string{}
			}
		} else {
			currentContent = append(currentContent, line)
		}
	}

	// Don't forget the last message
	if currentRole != "" && len(currentContent) > 0 {
		messages = append(messages, ChatMessage{
			Role:    strings.ToLower(currentRole),
			Content: strings.TrimSpace(strings.Join(currentContent, "\n")),
		})
	}

	return messages
}

// GetVariables extracts variable names from the prompt.
func (p *PromptVersion) GetVariables() []string {
	doubleBrace := regexp.MustCompile(`\{\{(\w+)\}\}`)
	singleBrace := regexp.MustCompile(`\{(\w+)\}`)

	variableSet := make(map[string]struct{})

	for _, match := range doubleBrace.FindAllStringSubmatch(p.Prompt, -1) {
		variableSet[match[1]] = struct{}{}
	}
	for _, match := range singleBrace.FindAllStringSubmatch(p.Prompt, -1) {
		variableSet[match[1]] = struct{}{}
	}

	variables := make([]string, 0, len(variableSet))
	for v := range variableSet {
		variables = append(variables, v)
	}
	return variables
}

// PromptCache manages prompt caching.
type PromptCache struct {
	cache    map[string]cachedPrompt
	mu       sync.RWMutex
	cacheTTL time.Duration
}

type cachedPrompt struct {
	prompt   *PromptVersion
	cachedAt time.Time
}

var (
	defaultPromptCache = &PromptCache{
		cache:    make(map[string]cachedPrompt),
		cacheTTL: time.Minute,
	}
)

// GetPromptOptions holds options for fetching a prompt.
type GetPromptOptions struct {
	Name     string
	Version  *int
	Label    string
	Fallback string
	CacheTTL time.Duration
}

// GetPrompt fetches a prompt from the server.
func GetPrompt(opts GetPromptOptions) (*PromptVersion, error) {
	client := GetGlobalClient()
	if client == nil {
		return getFallback(opts)
	}

	cacheKey := getCacheKey(opts)
	cacheTTL := opts.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = defaultPromptCache.cacheTTL
	}

	// Check cache
	defaultPromptCache.mu.RLock()
	if cached, ok := defaultPromptCache.cache[cacheKey]; ok {
		if time.Since(cached.cachedAt) < cacheTTL {
			defaultPromptCache.mu.RUnlock()
			return cached.prompt, nil
		}
	}
	defaultPromptCache.mu.RUnlock()

	// Fetch from API
	params := url.Values{}
	params.Set("name", opts.Name)
	if opts.Version != nil {
		params.Set("version", fmt.Sprintf("%d", *opts.Version))
	}
	if opts.Label != "" {
		params.Set("label", opts.Label)
	}

	apiURL := fmt.Sprintf("%s/api/public/prompts?%s", client.config.Host, params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return getFallback(opts)
	}

	req.Header.Set("Authorization", "Bearer "+client.config.APIKey)
	req.Header.Set("User-Agent", "agenttrace-go/0.1.0")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return getFallback(opts)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getFallback(opts)
	}

	var prompt PromptVersion
	if err := json.NewDecoder(resp.Body).Decode(&prompt); err != nil {
		return getFallback(opts)
	}

	// Update cache
	defaultPromptCache.mu.Lock()
	defaultPromptCache.cache[cacheKey] = cachedPrompt{
		prompt:   &prompt,
		cachedAt: time.Now(),
	}
	defaultPromptCache.mu.Unlock()

	return &prompt, nil
}

func getCacheKey(opts GetPromptOptions) string {
	parts := []string{opts.Name}
	if opts.Version != nil {
		parts = append(parts, fmt.Sprintf("v%d", *opts.Version))
	}
	if opts.Label != "" {
		parts = append(parts, "l:"+opts.Label)
	}
	return strings.Join(parts, ":")
}

func getFallback(opts GetPromptOptions) (*PromptVersion, error) {
	if opts.Fallback != "" {
		return &PromptVersion{
			ID:      "fallback",
			Version: 0,
			Prompt:  opts.Fallback,
			Labels:  []string{"fallback"},
		}, nil
	}
	return nil, fmt.Errorf("prompt '%s' not found and no fallback provided", opts.Name)
}

// SetPromptCacheTTL sets the default cache TTL for prompts.
func SetPromptCacheTTL(ttl time.Duration) {
	defaultPromptCache.mu.Lock()
	defer defaultPromptCache.mu.Unlock()
	defaultPromptCache.cacheTTL = ttl
}

// ClearPromptCache clears the prompt cache.
func ClearPromptCache() {
	defaultPromptCache.mu.Lock()
	defer defaultPromptCache.mu.Unlock()
	defaultPromptCache.cache = make(map[string]cachedPrompt)
}

// InvalidatePrompt invalidates cache for a specific prompt.
func InvalidatePrompt(name string) {
	defaultPromptCache.mu.Lock()
	defer defaultPromptCache.mu.Unlock()

	for key := range defaultPromptCache.cache {
		if strings.HasPrefix(key, name) {
			delete(defaultPromptCache.cache, key)
		}
	}
}
