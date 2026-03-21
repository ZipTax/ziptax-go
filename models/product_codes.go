package models

// ProductCodeSearchRequest is the request payload for product code search and
// recommendation endpoints. It contains a natural language product description
// used to find matching Taxability Information Codes (TICs).
type ProductCodeSearchRequest struct {
	Query string `json:"query"`
}

// ProductCodeSearchResult represents a single product code search result with
// TIC code, description, rank, and score.
//
// The API returns ticId, rank, and score as string values.
type ProductCodeSearchResult struct {
	// TicID is the Taxability Information Code. Use as the taxabilityCode
	// parameter in rate requests or cart line items.
	TicID string `json:"ticId"`

	// Label is the TIC label from the TIC data.
	Label string `json:"label"`

	// NaturalLabel is a natural language label aligned with the full description.
	NaturalLabel string `json:"naturalLabel"`

	// Description is the full description of the taxabilityCode line item.
	Description string `json:"description"`

	// Documentation is the long-form documentation of the TIC code.
	Documentation string `json:"documentation"`

	// Rank is the itemized rank for the result ("1" = best match).
	Rank string `json:"rank"`

	// Score is the confidence score ("0.0"-"1.0"), independent of rank.
	Score string `json:"score"`
}

// ProductCodeSearchResponse is the response from the product code search endpoint.
// It contains the original query and a list of matching product codes ranked
// and scored by relevance.
//
// Example:
//
//	response, err := client.SearchProductCodes(ctx, "baked goods")
//	if err != nil {
//		return err
//	}
//	for _, result := range response.Results {
//		fmt.Printf("TIC %s: %s (rank %s, score %s)\n",
//			result.TicID, result.Label, result.Rank, result.Score)
//	}
type ProductCodeSearchResponse struct {
	// Query is the original search query sent in the request.
	Query string `json:"query"`

	// Results contains matching product codes ranked by relevance.
	Results []ProductCodeSearchResult `json:"results"`
}

// ProductCodeRecommendation represents a single AI-powered product code
// recommendation. Check the Status field to determine if the prediction
// succeeded or failed.
type ProductCodeRecommendation struct {
	// Status is the prediction result status ("success" or "fail").
	Status string `json:"status"`

	// Error is a non-null error message when the prediction fails.
	// This field is only populated when Status is "fail".
	Error *string `json:"error"`

	// TicID is the recommended Taxability Information Code.
	TicID string `json:"ticId"`

	// Label is the TIC label from the TIC data.
	Label string `json:"label"`

	// NaturalLabel is a natural language label aligned with the description.
	NaturalLabel string `json:"naturalLabel"`

	// TicDescription is the full description of the recommended TIC code.
	// Note: JSON tag uses snake_case ("tic_description") to match the API response format,
	// unlike sibling fields which use camelCase. This is intentional.
	TicDescription string `json:"tic_description"`

	// ProductDescription is the original product description from the query.
	// Note: JSON tag uses snake_case ("product_description") to match the API response format,
	// unlike sibling fields which use camelCase. This is intentional.
	ProductDescription string `json:"product_description"`
}

// ProductCodeRecommendationResponse is the response from the product code
// recommendation endpoint. It contains AI-powered product code recommendations
// (typically one).
//
// Example:
//
//	response, err := client.RecommendProductCode(ctx, "baked goods")
//	if err != nil {
//		return err
//	}
//	prediction := response.Predictions[0]
//	if prediction.Status == "success" {
//		fmt.Printf("Recommended TIC: %s (%s)\n", prediction.TicID, prediction.Label)
//	}
type ProductCodeRecommendationResponse struct {
	// Predictions contains AI-powered product code recommendations.
	Predictions []ProductCodeRecommendation `json:"predictions"`
}
