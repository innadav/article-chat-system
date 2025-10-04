package prompts

import (
	"article-chat-system/internal/models"
	"bytes" // Use bytes.Buffer instead of strings.Builder for templates
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const ( // Define LLM models here
	ModelGemini15Flash = "gemini-1.5-flash"
)

type PromptTemplate struct {
	Template string `yaml:"template"`
}

type Loader struct {
	PromptDir string
	Cache     map[string]*template.Template
}

func NewLoader(version string) (*Loader, error) {
	promptDir := filepath.Join("configs", "prompts", version)
	return &Loader{
		PromptDir: promptDir,
		Cache:     make(map[string]*template.Template),
	}, nil
}

func (l *Loader) LoadPrompt(name string) (*template.Template, error) {
	if t, ok := l.Cache[name]; ok {
		return t, nil
	}

	filePath := filepath.Join(l.PromptDir, name+".yaml")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}

	var pt PromptTemplate
	if err := yaml.Unmarshal(data, &pt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompt YAML from %s: %w", filePath, err)
	}

	t, err := template.New(name).Parse(pt.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	l.Cache[name] = t
	return t, nil
}

type Factory struct {
	Model  string
	Loader *Loader
}

func NewFactory(model string, loader *Loader) (*Factory, error) { // Return error from constructor
	// It's good practice to pre-load/parse templates here if possible
	return &Factory{Model: model, Loader: loader}, nil
}

func (f *Factory) executeTemplate(name string, data interface{}) (string, error) {
	tmpl, err := f.Loader.LoadPrompt(name)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

// --- FIX: CreatePlannerPrompt now uses the external template ---
func (f *Factory) CreatePlannerPrompt(query string, articleTitles []*models.Article) (string, error) {
	var titles []string
	for _, art := range articleTitles {
		titles = append(titles, art.Title)
	}
	data := struct {
		Query    string
		Articles string
	}{
		Query:    query,
		Articles: strings.Join(titles, "\n- "),
	}
	return f.executeTemplate("planner", data)
}

// CreateSummarizePrompt generates a prompt for summarizing an article.
func (f *Factory) CreateSummarizePrompt(articleContent string) (string, error) {
	data := struct{ Content string }{Content: articleContent}
	return f.executeTemplate("summarize", data)
}

// CreateKeywordsPrompt generates a prompt for extracting keywords.
func (f *Factory) CreateKeywordsPrompt(articleTitle, articleContent string) (string, error) {
	data := struct{ Title, Content string }{Title: articleTitle, Content: articleContent}
	return f.executeTemplate("keywords", data)
}

// CreateInitialAnalysisPrompt generates a prompt for comprehensive initial analysis.
func (f *Factory) CreateInitialAnalysisPrompt(content string) (string, error) {
	data := struct{ Content string }{Content: content}
	return f.executeTemplate("initial_analysis", data)
}

// CreateEntityExtractionPrompt generates a prompt for comprehensive entity extraction.
func (f *Factory) CreateEntityExtractionPrompt(title, excerpt string) (string, error) {
	data := struct{ Title, Excerpt string }{Title: title, Excerpt: excerpt}
	return f.executeTemplate("entity_extraction", data)
}

// CreateFindTopicPrompt generates a prompt for finding and synthesizing articles about a topic
func (f *Factory) CreateFindTopicPrompt(topic string, articles []*models.Article) (string, error) {
	data := map[string]interface{}{
		"Topic":    topic,
		"Articles": articles,
	}
	return f.executeTemplate("find_topic", data)
}

// CreateComparePositivityPrompt generates a prompt for comparing positivity across articles
func (f *Factory) CreateComparePositivityPrompt(topic string, articles []*models.Article) (string, error) {
	data := map[string]interface{}{
		"Topic":    topic,
		"Articles": articles,
	}
	return f.executeTemplate("compare_positivity", data)
}
