package intelligence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/llm"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/openai/openai-go/v2/shared"
	lP "github.com/vertexcover-io/locatr/pkg"
	lLlm "github.com/vertexcover-io/locatr/pkg/llm"
	"github.com/vertexcover-io/locatr/pkg/mode"
	"github.com/vertexcover-io/locatr/pkg/plugins"
	"github.com/vertexcover-io/locatr/pkg/types"
)

var ErrLocatrCreationFailed = errors.New("failed to create locatr client")

const descriptionGenerationPrompt = `
You are an AI agent that is responsible for describing an element to a user. You are an experienced UX developer that is able to understand the
role a given element has on the DOM. You are more concerned with the relationship of the element with its surrounding that it's attribute, that you
take as suplemental information. You are capable of understanding which of the sibling, child or parent element best describes the element. If the parent of the parent element describes the section better then you will also include them. You will infact go as far above or lower as needed until you get a concrete description for uniqueness without depending on coordinates. You are to strictly avoid any xy coordinates or width and height ofr uniqueness of element. The description for the element is anchored to only those element that play a role in the action that the element takes upon the DOM. Your description
should be consciece and information dense but shouldn't exclude any significant information either. A user upon reading the element must be able to identify the element in question from the DOM, even if they are provided with similar DOM with only the text of the element differing.
Your output should be a json object, with 2 values 'reason' and 'summary'. 'reason' is where you put your thoughts on why you think that the
summary you have generated best fits the scenario. and 'summary' is where you put your actual description. Also scan the entire DOM once you have determined the element, if you find multiple similar element, then you need to include what makes the particular element unique among it's peers.
Also add the following line to the summary, "The user requirement may not match exactly. Make sure you account for synonyms when finding the
element. Also utilize other information available in the requirement, if you don't find the exact matches."
`

const iosScreenshotComparePrompt = `
You are tasked with analyzing mobile app screenshots to determine if they belong to the same page, different sections of the same page, or entirely different pages. Use the following criteria to make your assessment:

Page Title and Header Consistency:

Check for consistent titles or headers. Identical or similar headers often indicate the same page or type of page.
Visual and Layout Consistency:

Observe the overall layout, including navigation buttons, icons, and menu styles. Consistent layouts suggest the same page type.
Content and Functional Similarity:

Identify if the content serves a similar function or purpose (e.g., playlists, settings categories, user profiles). Consistent functionality indicates the same page type.
Styling and Design Elements:

Ensure the font, color scheme, and iconography are uniform across screenshots. Similar design elements often indicate the same page.
Contextual Clues and Features:

Look for shared features, such as buttons or options, that connect the screenshots as part of the same page type.
Dynamic Content Recognition:

Recognize that variations in dynamic content (e.g., playlist names, product listings) do not necessarily indicate different pages if other elements remain consistent.
Use these characteristics to determine if the screenshots depict the same page or type of page. 

Your response should be in JSON format with 2 properties. The first property named "reason", gives a short feedback on your decision, while the second
property called "result", gives boolean value if it is same or not.
`

type llmIntelSystem struct {
	llm    llm.LLM
	apiKey string
}

func NewLLmIntelSystem(llm llm.LLM, apiKey string) *llmIntelSystem {
	return &llmIntelSystem{
		llm:    llm,
		apiKey: apiKey,
	}
}

func (l *llmIntelSystem) GenerateElementDescription(ctx context.Context, root, elem string) (string, error) {
	c := fmt.Sprintf(`
	 =============== DOM ==============
	%s
	==================================

	================ Element ==========
	%s
	===================================
	`, root, elem)

	m := []llm.Message{
		{
			Role: llm.SystemRole,
			Content: []llm.MessageContent{
				{
					Type: llm.TextContent,
					Data: descriptionGenerationPrompt,
				},
			},
		},
		{
			Role: llm.UserRole,
			Content: []llm.MessageContent{
				{
					Type: llm.TextContent,
					Data: c,
				},
			},
		},
	}

	resp, err := l.llm.Completion(ctx, m, shared.ChatModelO3Mini, true)
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrDescriptionGenerationFailed)
	}

	p := struct {
		Summary string `json:"summary,omitempty"`
	}{}
	if err := json.Unmarshal([]byte(resp), &p); err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrDescriptionGenerationFailed)
	}

	return p.Summary, nil
}

func (l *llmIntelSystem) CompareScreenShot(ctx context.Context, img1, img2 string) (bool, error) {
	m := []llm.Message{
		{
			Role: llm.SystemRole,
			Content: []llm.MessageContent{
				{
					Type: llm.TextContent,
					Data: iosScreenshotComparePrompt,
				},
			},
		},
		{
			Role: llm.UserRole,
			Content: []llm.MessageContent{
				{
					Type: llm.TextContent,
					Data: "Here are the screenshots",
				},
				{
					Type:   llm.ImageContent,
					Data:   img1,
					Detail: "low",
				},
				{
					Type:   llm.ImageContent,
					Data:   img2,
					Detail: "low",
				},
			},
		},
	}

	resp, err := l.llm.Completion(ctx, m, shared.ChatModelGPT4o2024_11_20, true)
	if err != nil {
		return false, fmt.Errorf("%w: %w", err, ErrSSComparisionFailed)
	}

	p := struct {
		Reason string `json:"reason,omitempty"`
		Result bool   `json:"result,omitempty"`
	}{}
	if err := json.Unmarshal([]byte(resp), &p); err != nil {
		return false, fmt.Errorf("%w: %w", err, ErrSSComparisionFailed)
	}

	return p.Result, nil
}

func (l *llmIntelSystem) GenerateLocator(ctx context.Context, desc string, root page.Page, platform platform.Platform) (string, error) {
	locatr, err := createLocatr(locatrOpts{
		apiKey:   l.apiKey,
		page:     root,
		pageType: pageTypeConverter(root.PageType()),
		platform: platformConverter(platform),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrNewLocatorGenerationFailed)
	}

	comp, err := locatr.Locate(ctx, desc)
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, ErrNewLocatorGenerationFailed)
	}

	if len(comp.Locators) < 1 {
		return "", fmt.Errorf("failed to generate any valid locators: %w", ErrNewLocatorGenerationFailed)
	}
	lctr := comp.Locators[0]

	if comp.LocatorType == types.CssSelectorType {
		lctr = page.ConvertCssSelectorToXpath(lctr)
	}
	return lctr, err
}

type locatrOpts struct {
	apiKey   string
	page     page.Page
	pageType plugins.PageType
	platform plugins.Platform
}

func createLocatr(opts locatrOpts) (*lP.Locatr, error) {

	plugin := plugins.NewRawTextPlugin(opts.page.String(), opts.pageType, opts.platform)

	llmClient, err := lLlm.NewLLMClient(
		lLlm.WithProvider(lLlm.OpenAI),
		lLlm.WithAPIKey(opts.apiKey),
		lLlm.WithModel(shared.ChatModelGPT4oMini),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", err, ErrLocatrCreationFailed)
	}

	locatr, err := lP.NewLocatr(plugin,
		lP.WithLLMClient(llmClient),
		lP.WithRerankerDisabled(),
		lP.WithMode(&mode.DOMAnalysisMode{}),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", err, ErrLocatrCreationFailed)
	}

	return locatr, nil
}

func pageTypeConverter(p page.PageType) plugins.PageType {
	if p == page.HTMLPageType {
		return plugins.HTMLPageType
	}
	return plugins.XMLPageType
}

func platformConverter(p platform.Platform) plugins.Platform {
	switch p {
	case platform.AndroidPlatform:
		return plugins.AndroidPlatform
	case platform.IosPlatform:
		return plugins.IosPlatform
	}
	return plugins.WebPlatform
}
