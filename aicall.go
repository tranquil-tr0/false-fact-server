package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Request structure for AI API
type AnalyzeArticleRequest struct {
	Content    string    `json:"content"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	LastEdited time.Time `json:"last_edited"`
}

// Response structure for AI API
type Reasoning struct {
	Factual    []string `json:"factual"`
	Unfactual  []string `json:"unfactual"`
	Subjective []string `json:"subjective"`
	Objective  []string `json:"objective"`
}

type Categories struct {
	Factuality  int `json:"factuality"`
	Objectivity int `json:"objectivity"`
}

type AnalyzeArticleResponse struct {
	Reasoning        Reasoning  `json:"reasoning"`
	CredibilityScore int        `json:"credibilityScore"`
	Categories       Categories `json:"categories"`
	Confidence       int        `json:"confidence"`
	Sources          []string   `json:"sources"`
}

// Calls the external AI API for article analysis
func CallAIAnalyzeArticle(content string, title string, url string, lastEdited time.Time) (*AnalyzeArticleResponse, error) {
	aiAPIURL := "https://text.pollinations.ai/openai"

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
  "credibilityScore": <number 0-100>,  "categories": {
    "factuality": <percentage 0-100>,
    "objectivity": <percentage 0-100>
  },
  "confidence": <number 0-100>
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

Your response must be in the format specified above.
`

	payload := map[string]interface{}{
		"model": "openai-fast",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": analysisPrompt},
		},
		"temperature":     0.7,
		"stream":          false,
		"private":         false,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", aiAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("AI API returned non-200 status")
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, errors.New("No choices returned from AI API")
	}

	// The content field should be a JSON string matching AnalyzeArticleResponse
	var result AnalyzeArticleResponse
	if err := json.Unmarshal([]byte(apiResp.Choices[0].Message.Content), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

geminiApiCall ()
pollinationsApiCall ()