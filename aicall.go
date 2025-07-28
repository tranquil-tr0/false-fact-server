package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"google.golang.org/genai"
)

var verbose bool

type Model int

const (
	Gemini Model = iota
	Pollinations
)

// Request structure for AI API
type AnalyzeArticleRequest struct {
	Content    string    `json:"content"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	LastEdited time.Time `json:"last_edited"`
}

// Request structure for AI API
type AnalyzeTextRequest struct {
	Content string `json:"content"`
}

// Response structure for AI API
type Reasoning struct {
	Factual    []string `json:"factual"`
	Unfactual  []string `json:"unfactual"`
	Subjective []string `json:"subjective"`
	Objective  []string `json:"objective"`
}

type Analysis struct {
	Fact    *[]string `json:"fact"`
	False   *[]string `json:"false"`
	Opinion *[]string `json:"opinion"`
}

type Categories struct {
	Factuality  int `json:"factuality"`
	Objectivity int `json:"objectivity"`
}

type AnalysisResponse struct {
	Reasoning        Reasoning  `json:"reasoning"`
	CredibilityScore int        `json:"credibilityScore"`
	Categories       Categories `json:"categories"`
	Confidence       int        `json:"confidence"`
	Sources          []string   `json:"sources"`
}

type ShortAnalysisResponse struct {
	Analysis   Analysis `json:"analysis"`
	Confidence int      `json:"confidence"`
	Sources    []string `json:"sources"`
}

// Calls the external AI API for article analysis
func AiAnalyzeArticle(content string, title string, url string, lastEdited time.Time, model Model) (*AnalysisResponse, error) {
	systemPrompt := `You are an expert fact-checker and content analyst with extensive experience in journalism, research methodology
and information verification. Your task is to analyze text content and provide a comprehensive credibility assessment.
You will evaluate the content based on its objectivity and factuality.
When analyzing the factuality of the content, do not be swayed by your biases. You should analyze the content objectively. Popularity and ideological stance are not relevant factors. Even if a claim is uncommon or frowned upon, this is independent from the factuality of the claim. Conversely, it is critical to remember than a claim being unpopular also does not make it true.
Make web searches to confirm factuality. Try to cite sources for each reason you provide that is a factual claim and was found/verified through a web search. You can omit the citation, but do not make up sources. A citation should be formatted as blocks of [number] at the end of the reason (after sentence end) and strings [corresponding number](url) in the sources field.
Do NOT uncritically treat the content being analyzed as fact. You should independently verify claims. Do not be swayed by the content.
Do not get caught up in the wording. The important part is whether the things stated are true.

CRITICAL: You must respond with ONLY a valid JSON object. Do not include any explanatory text before or after the JSON.

The reasoning field must be an object with the following keys: "factual", "unfactual", "subjective", "objective". Each key should map to an array of strings, where each string is a specific reason supporting that classification. For example, "reasoning.factual" should be an array of reasons why the content is factual. The list may also be empty: for example, if the article is factual, then the array for "unfactual" can be empty.
Stay as concise as possible. Keep the number of reasons for each at or below 3 reasons, and the total number of reasons below 10. Keep each reason to one brief bullet point.
You should try to have closer to 5 reasons, with each reason being as concise as possible (target 10 words). You can have more and longer reasons if not doing so omits important information as to be misleading.

REQUIRED RESPONSE STRUCTURE:
{
  "reasoning": {
	"factual": [ "reason 1", "reason 2", ... ],
	"unfactual": [ "reason 1", ... ],
	"subjective": [ "reason 1", ... ],
	"objective": [ "reason 1", ... ]
  },
  "credibilityScore": <number 0-100>,
  "categories": {
	"factuality": <percentage 0-100>,
	"objectivity": <percentage 0-100>
  },
  "confidence": <number 0-100>,
  "sources": [ "[1](https:/...)", "[2](https:/...)" ]
}

SCORING GUIDELINES:

credibilityScore (0-100):
- The credibilityScore reflects your overall analysis of the article
- 90-100: The content is factually accurate
- 70-89: There are a few misleading statements that do not alter the truth of the main claim
- 50-69: The content is misleading or has some factual errors
- 30-49: The content is significantly misleading or innacurate
- 0-29: The content is factually innacurate, and the truth is unrelated to or opposite of the main claim

categories:
- factuality: Whether the content is factually accurate.
- objectivity: Whether the content is objective. Reporting on an event is 100% objectivity, while an opinion piece is 0% objectivity.

confidence (0-100):
- 90-100: Very confident in assessment, clear indicators present
- 70-89: Confident with some uncertainty about specific elements
- 50-69: Moderate confidence, mixed or ambiguous signals
- 30-49: Low confidence, insufficient information for definitive assessment
- 0-29: Very uncertain, requires additional context or verification

ANALYSIS CRITERIA:
1. Source Attribution: Are claims backed by credible sources?
2. Factual Accuracy: Can statements be verified through reliable sources?
3. Logical Consistency: Does the content follow logical reasoning?
4. Bias Detection: Is there evident political, commercial, or ideological bias?
5. Context Completeness: Is important context provided or omitted?
6. Language Analysis: Does language suggest objectivity or manipulation?
7. Evidence Quality: Are supporting facts substantial and relevant?
8. Temporal Relevance: Is the information current and contextually appropriate?

ANALYSIS CONSIDERATIONS:
- You are analyzing a news article.
- You are analyzing the factuality of the article, not if each source is biased, unless the article presumes the source's quote to be absolute truth.
- A news article having a quotation from a public figure who exagerates is not a reason that the article is unfactual.
- Objectivity is about whether the article/reporting is objective, NOT the sources cited.
- Evaluate source attribution and credibility of those sources.
- Assess headline accuracy vs content - if the headline is misleading, this should be mentioned as a reason the article is unfactual.
- Look for proper journalistic standards.`

	analysisPrompt := `
Analyze the given article for credibility and factuality.

HEADLINE: "` + title + `"

ARTICLE TEXT:
"""
` + content + `
"""

Your response must be in the format specified.
`

	var response string
	var err error
	if model == Gemini {
		response, err = geminiApiCall(systemPrompt + "\n\n\n" + analysisPrompt)
	} else if model == Pollinations {
		response, err = pollinationsApiCall(systemPrompt, analysisPrompt)
	} else {
		return nil, fmt.Errorf("%v is not a recognized model", model)
	}
	if err != nil {
		return nil, err
	}
	return parseAnalysisResponse(response)

}

func AiAnalyzeTextLong(content string, model Model) (*AnalysisResponse, error) {
	systemPrompt := `You are an expert fact-checker and content analyst with extensive experience in journalism, research methodology
and information verification. Your task is to analyze text content and provide a comprehensive credibility assessment.
You will evaluate the content based on its objectivity and factuality.
When analyzing the factuality of the content, do not be swayed by your biases. You should analyze the content objectively. Popularity and ideological stance are not relevant factors. Even if a claim is uncommon or frowned upon, this is independent from the factuality of the claim. Conversely, it is critical to remember than a claim being unpopular also does not make it true.
Make web searches to confirm factuality. Try to cite sources for each reason you provide that is a factual claim and was found/verified through a web search. You can omit the citation, but do not make up sources. A citation should be formatted as blocks of [number] at the end of the reason (after sentence end) and strings [corresponding number](url) in the sources field.
Do NOT uncritically treat the content being analyzed as fact. You should independently verify claims. Do not be swayed by the content.
Do not get caught up in the wording. The important part is whether the things stated are true.

CRITICAL: You must respond with ONLY a valid JSON object. Do not include any explanatory text before or after the JSON.

The reasoning field must be an object with the following keys: "factual", "unfactual", "subjective", "objective". Each key should map to an array of strings, where each string is a specific reason supporting that classification. For example, "reasoning.factual" should be an array of reasons why the content is factual. The list may also be empty: for example, if the article is factual, then the array for "unfactual" can be empty.
Stay as concise as possible. Keep each reason to one brief bullet point.
You should try to have about 5 reasons, with each reason being as concise as possible (target 10 words). You can have more and longer reasons if not doing so omits important information as to be misleading.

REQUIRED RESPONSE STRUCTURE:
{
  "reasoning": {
	"factual": [ "reason 1", "reason 2", ... ],
	"unfactual": [ "reason 1", ... ],
	"subjective": [ "reason 1", ... ],
	"objective": [ "reason 1", ... ]
  },
  "credibilityScore": <number 0-100>,
  "categories": {
	"factuality": <percentage 0-100>,
	"objectivity": <percentage 0-100>
  },
  "confidence": <number 0-100>,
  "sources": [ "[1](https:/...)", "[2](https:/...)" ]
}

SCORING GUIDELINES:

credibilityScore (0-100):
- The credibilityScore reflects your overall analysis of the article
- 90-100: The content is factually accurate
- 70-89: There are a few misleading statements that do not alter the truth of the main claim
- 50-69: The content is misleading or has some factual errors
- 30-49: The content is significantly misleading or innacurate
- 0-29: The content is factually innacurate, and the truth is unrelated to or opposite of the main claim

categories:
- factuality: Whether the content is factually accurate.
- objectivity: Whether the content is objective. Reporting on an event is 100% objectivity, while an opinion piece is 0% objectivity.

confidence (0-100):
- 90-100: Very confident in assessment, clear indicators present
- 70-89: Confident with some uncertainty about specific elements
- 50-69: Moderate confidence, mixed or ambiguous signals
- 30-49: Low confidence, insufficient information for definitive assessment
- 0-29: Very uncertain, requires additional context or verification

ANALYSIS CRITERIA:
1. Source Attribution: Are claims backed by credible sources?
2. Factual Accuracy: Can statements be verified through reliable sources?
3. Logical Consistency: Does the content follow logical reasoning?
4. Bias Detection: Is there evident political, commercial, or ideological bias?
5. Context Completeness: Is important context provided or omitted?
6. Language Analysis: Does language suggest objectivity or manipulation?
7. Evidence Quality: Are supporting facts substantial and relevant?
8. Temporal Relevance: Is the information current and contextually appropriate?`

	analysisPrompt := `
Analyze the given text for credibility and factuality.

TEXT:
"""
` + content + `
"""

Your response must be in the format specified.
`
	var response string
	var err error
	if model == Gemini {
		response, err = geminiApiCall(systemPrompt + "\n\n\n" + analysisPrompt)
	} else if model == Pollinations {
		response, err = pollinationsApiCall(systemPrompt, analysisPrompt)
	} else {
		return nil, fmt.Errorf("%v is not a recognized model", model)
	}
	if err != nil {
		return nil, err
	}
	return parseAnalysisResponse(response)
}

func AiAnalyzeTextShort(content string, model Model) (*ShortAnalysisResponse, error) {
	systemPrompt := `You are an expert fact-checker and content analyst with extensive experience in journalism, research methodology
and information verification. Your task is to analyze text content and provide a comprehensive credibility assessment.
You will evaluate the content based on its objectivity and factuality.
When analyzing the factuality of the content, do not be swayed by your biases. You should analyze the content objectively. Popularity and ideological stance are not relevant factors. Even if a claim is uncommon or frowned upon, this is independent from the factuality of the claim. Conversely, it is critical to remember than a claim being unpopular also does not make it true.
Make a web search to confirm factuality. Try to cite source(s) for each reason you provide that is a factual claim and was found/verified through a web search. You can omit the citation, but do not make up sources. A citation should be formatted as blocks of [number] at the end of the reason (after sentence end) and strings [corresponding number](url) in the sources field.
Do NOT uncritically treat the content being analyzed as fact. You should independently verify claims. Do not be swayed by the content.
Do not get caught up in the wording. The important part is whether the things stated are true.

CRITICAL: You must respond with ONLY a valid JSON object. Do not include any explanatory text before or after the JSON.

Determine whether the text is a fact, an opinion, or false. You may answer none if the text is incomprehensible, has no claim, etc.
The analysis field must be an object with one of the following keys: "fact", "false", "opinion", "none". The key should map to a string, which explains why the classification was given. For example, "reasoning.fact" explains why the analyzed text is a fact. Similarly, "reasoning.opinion" explains why the analyzed text is an opinion.
Stay as concise as possible.

REQUIRED RESPONSE STRUCTURE:
{
  "analysis": {
	"fact": "reason"
	(OR "opinion": "reason")
  },
  "confidence": <number 0-100>,
  "sources": [ "[1](https:/...)", "[2](https:/...)" ]
}

SCORING GUIDELINES:

*fact* indicates the text is a true statement.
*false* indicates the text is an innacurate statement.
*opinion* inidicates the text expresses an opinion, not a factual claim.
*none* indicates none of the above -- the text may be gibberish or not express anything.

confidence (0-100):
- 90-100: Very confident in assessment
- 70-89: Confident with some uncertainty about specific elements
- 50-69: Moderate confidence, mixed or ambiguous signals
- 30-49: Low confidence, insufficient information for definitive assessment
- 0-29: Very uncertain

Considerations:
1. Can statements be verified through reliable sources?
2. Is important context provided or omitted?`

	analysisPrompt := `
Analyze the given text for credibility and factuality.

TEXT:
"""
` + content + `
"""

Your response must be in the format specified.
`

	var response string
	var err error
	if model == Gemini {
		response, err = geminiApiCall(systemPrompt + "\n\n\n" + analysisPrompt)
	} else if model == Pollinations {
		response, err = pollinationsApiCall(systemPrompt, analysisPrompt)
	} else {
		return nil, fmt.Errorf("%v is not a recognized model", model)
	}
	if err != nil {
		return nil, err
	}
	return parseShortAnalysisResponse(response)
}

func geminiApiCall(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if len(apiKey) == 0 {
		return "", &ExtensionError{
			Type:        ApiUnavailable,
			Message:     "Gemini API key is missing",
			Retryable:   false,
			UserMessage: "Please set GEMINI_API_KEY in your environment",
		}
	}

	if verbose {
		fmt.Printf("[Gemini] Using prompt: %s\n", prompt)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", &ExtensionError{
			Type:        ApiUnavailable,
			Message:     "Failed to initialize Gemini client: " + err.Error(),
			Retryable:   true,
			UserMessage: err.Error(),
		}
	}

	modelName := "gemini-2.5-flash"
	temperature := genai.Ptr[float32](0.5)
	thinkingBudget := int32(0) // disables thinking

	result, err := client.Models.GenerateContent(
		ctx,
		modelName,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			Temperature: temperature,
			ThinkingConfig: &genai.ThinkingConfig{
				ThinkingBudget: &thinkingBudget,
			},
			Tools: []*genai.Tool{
				{GoogleSearch: &genai.GoogleSearch{}},
			},
		},
	)
	if err != nil {
		return "", &ExtensionError{
			Type:        ApiUnavailable,
			Message:     "Gemini API request failed: " + err.Error(),
			Retryable:   true,
			UserMessage: err.Error(),
		}
	}

	content := ""
	if result != nil {
		content = result.Text()
		if verbose {
			fmt.Printf("[Gemini] Received content: %s\n", content)
		}
	}

	return content, nil
}

func pollinationsApiCall(systemPrompt string, userPrompt string) (string, error) {
	payload := map[string]interface{}{
		"model": "openai-fast",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature":     0.7,
		"stream":          false,
		"private":         false,
		"response_format": map[string]string{"type": "json_object"},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Printf("[Pollinations] Sending payload: %s\n", string(payloadBytes))
	}

	req, err := http.NewRequest("POST", "https://text.pollinations.ai/openai", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", handleHttpStatusError(resp.StatusCode, fmt.Sprintf("POST request failed with status %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Printf("[Pollinations] Received response body: %s\n", string(body))
	}

	var responseJson map[string]interface{}
	if err := json.Unmarshal(body, &responseJson); err != nil {
		return "", err
	}

	var content string
	if choices, ok := responseJson["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if c, ok := message["content"].(string); ok {
					content = c
					if verbose {
						fmt.Printf("[Pollinations] Extracted content: %s\n", content)
					}
				}
			}
		}
	}

	return content, nil
}

func parseAnalysisResponse(content string) (*AnalysisResponse, error) {
	if verbose {
		fmt.Printf("[Parse] Raw content for parsing: %s\n", content)
	}
	content = string(bytes.TrimSpace([]byte(content)))

	// Extract first JSON object from the response
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	jsonMatch := re.FindString(content)
	if jsonMatch != "" {
		content = jsonMatch
		if verbose {
			fmt.Printf("[Parse] Extracted JSON: %s\n", content)
		}
	}

	var parsed AnalysisResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		if verbose {
			fmt.Printf("[Parse] Failed to unmarshal: %v\n", err)
		}
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Failed to parse analysis response",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}

	// Validate required fields
	if parsed.CredibilityScore == 0 && parsed.Categories.Factuality == 0 && parsed.Categories.Objectivity == 0 &&
		parsed.Confidence == 0 && len(parsed.Reasoning.Factual) == 0 && len(parsed.Reasoning.Unfactual) == 0 &&
		len(parsed.Reasoning.Subjective) == 0 && len(parsed.Reasoning.Objective) == 0 {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Invalid response format from analysis service",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}

	// Validate score ranges
	if parsed.CredibilityScore < 0 || parsed.CredibilityScore > 100 {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Invalid credibility score in response",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}
	if parsed.Confidence < 0 || parsed.Confidence > 100 {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Invalid confidence score in response",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}
	// Validate categories
	if parsed.Categories.Factuality < 0 || parsed.Categories.Factuality > 100 ||
		parsed.Categories.Objectivity < 0 || parsed.Categories.Objectivity > 100 {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Category values out of range",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}
	// Validate sources
	if parsed.Sources == nil {
		parsed.Sources = []string{}
	}
	filteredSources := []string{}
	for _, s := range parsed.Sources {
		if len(s) > 0 {
			filteredSources = append(filteredSources, s)
		}
	}
	parsed.Sources = filteredSources

	return &parsed, nil
}

func parseShortAnalysisResponse(content string) (*ShortAnalysisResponse, error) {
	if verbose {
		fmt.Printf("[Parse] Raw content for parsing: %s\n", content)
	}
	content = string(bytes.TrimSpace([]byte(content)))

	// Extract first JSON object from the response
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	jsonMatch := re.FindString(content)
	if jsonMatch != "" {
		content = jsonMatch
		if verbose {
			fmt.Printf("[Parse] Extracted JSON: %s\n", content)
		}
	}

	var parsed ShortAnalysisResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		if verbose {
			fmt.Printf("[Parse] Failed to unmarshal: %v\n", err)
		}
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Failed to parse analysis response",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}

	if parsed.Confidence < 0 || parsed.Confidence > 100 {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Invalid confidence score in response",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}

	if (parsed.Analysis.Fact != nil && parsed.Analysis.False != nil) ||
		(parsed.Analysis.Fact != nil && parsed.Analysis.Opinion != nil) ||
		(parsed.Analysis.False != nil && parsed.Analysis.Opinion != nil) {
		return nil, &ExtensionError{
			Type:        InvalidContent,
			Message:     "Multiple analysis conclusions",
			Retryable:   true,
			UserMessage: "Try analyzing the content again",
		}
	}

	// Validate sources
	if parsed.Sources == nil {
		parsed.Sources = []string{}
	}
	filteredSources := []string{}
	for _, s := range parsed.Sources {
		if len(s) > 0 {
			filteredSources = append(filteredSources, s)
		}
	}
	parsed.Sources = filteredSources

	return &parsed, nil
}

type AnalysisErrorType string

const (
	RateLimited    AnalysisErrorType = "RATE_LIMITED"
	ApiUnavailable AnalysisErrorType = "API_UNAVAILABLE"
	InvalidContent AnalysisErrorType = "INVALID_CONTENT"
	NetworkError   AnalysisErrorType = "NETWORK_ERROR"
)

type ExtensionError struct {
	Type        AnalysisErrorType
	Message     string
	Retryable   bool
	UserMessage string
}

func (e *ExtensionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func handleHttpStatusError(status int, message string) error {
	switch {
	case status == 429:
		return &ExtensionError{
			Type:        RateLimited,
			Message:     "API rate limit exceeded",
			Retryable:   true,
			UserMessage: "Please wait a moment before trying again",
		}
	case status >= 500:
		return &ExtensionError{
			Type:        ApiUnavailable,
			Message:     "Analysis service is temporarily unavailable",
			Retryable:   true,
			UserMessage: "Please try again in a few minutes",
		}
	case status == 404:
		return &ExtensionError{
			Type:        ApiUnavailable,
			Message:     "API endpoint not found (404)",
			Retryable:   false,
			UserMessage: "Using fallback analysis method",
		}
	case status == 400:
		return &ExtensionError{
			Type:        InvalidContent,
			Message:     "Invalid request format or content",
			Retryable:   false,
			UserMessage: "Please try with different content or check your input",
		}
	case status >= 400 && status < 500:
		return &ExtensionError{
			Type:        ApiUnavailable,
			Message:     fmt.Sprintf("API request failed with status %d", status),
			Retryable:   false,
			UserMessage: "Please check your request and try again",
		}
	default:
		return &ExtensionError{
			Type:        NetworkError,
			Message:     message,
			Retryable:   true,
			UserMessage: "Please check your internet connection and try again",
		}
	}
}
