package mjml

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/Boostport/mjml-go"
	"github.com/aymerick/raymond"
)

//go:embed templates
var templates embed.FS

type MJMLService interface {
	GenerateHTMLFromMJML(ctx context.Context, template string, args map[string]string) (*string, error)
}

type mjmlService struct{}

func NewMJMLService() MJMLService {
	return &mjmlService{}
}

func (s *mjmlService) GenerateHTMLFromMJML(ctx context.Context, template string, args map[string]string) (*string, error) {
	// Get file source from embed.FS
	file, err := templates.ReadFile("templates/" + template)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", template, err)
	}

	// Handle mj-include tags
	expandedTemplate, err := s.handleIncludingTemplates(string(file))
	if err != nil {
		return nil, fmt.Errorf("failed to process includes: %w", err)
	}

	// Parse MJML
	output, err := mjml.ToHTML(ctx, expandedTemplate, mjml.WithMinify(true))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MJML: %w", err)
	}

	// Render the template with provided arguments
	result, err := raymond.Render(output, args)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return &result, nil
}

// handleIncludingTemplates replaces all mj-include tags with their content
func (s *mjmlService) handleIncludingTemplates(template string) (string, error) {
	var buffer bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(template))

	for scanner.Scan() {
		line := scanner.Text()

		// Check for mj-include tag
		if strings.Contains(line, "<mj-include") {
			// Extract file path from the mj-include tag
			includePath := extractIncludePath(line)
			if includePath == "" {
				return "", fmt.Errorf("failed to parse mj-include tag: %s", line)
			}

			// Remove `./` from the path
			includePath = strings.Replace(includePath, "./", "", 1)
			includePath = strings.TrimSpace(includePath)
			includeContent, err := templates.ReadFile("templates/" + includePath)
			if err != nil {
				return "", fmt.Errorf("failed to read included template %s: %w", includePath, err)
			}

			// Recursively process the included content
			processedInclude, err := s.handleIncludingTemplates(string(includeContent))
			if err != nil {
				return "", fmt.Errorf("failed to process included template %s: %w", includePath, err)
			}

			buffer.WriteString(processedInclude)
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading template: %w", err)
	}

	return buffer.String(), nil
}

// extractIncludePath parses the mj-include tag and extracts the file path
func extractIncludePath(line string) string {
	start := strings.Index(line, "path=\"")
	if start == -1 {
		return ""
	}
	start += len("path=\"")
	end := strings.Index(line[start:], "\"")
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}
