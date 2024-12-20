package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Auth represents a client for interacting with the Auth API.
type Auth struct {
	Domain       string       // Your Okta domain (e.g., "your-domain.okta.com").
	APIToken     string       // Your Okta API token.
	ClientID     string       // Your Okta application's client ID.
	ClientSecret string       // Your Okta application's client secret.
	HTTPClient   *http.Client // The HTTP client to use for API requests.
}

// New creates a new Okta client.
//
// Parameters:
//   - domain: Your Okta domain.
//   - apiToken: Your Okta API token.
//   - clientID: Your Okta application's client ID.
//   - clientSecret: Your Okta application's client secret.
//
// Returns:
//   - A new Okta client instance.
func New(domain, apiToken, clientID, clientSecret string) *Auth {
	return &Auth{
		Domain:       domain,
		APIToken:     apiToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second, // Default timeout of 10 seconds for API requests.
		},
	}
}

// RegistrationRequest represents the data needed to register a new user.
type RegistrationRequest struct {
	Profile   UserProfile `json:"profile"`   // The user's profile information.
	Activate  bool        `json:"activate"`  // Whether to activate the user immediately.
	SendEmail bool        `json:"sendEmail"` // Whether to send activation email if not activated immediately.
	SendSMS   bool        `json:"sendSMS"`   // Whether to send activation SMS if not activated immediately.
}

// UserProfile represents the user's profile information.
type UserProfile struct {
	FirstName   string `json:"firstName"`             // The user's first name.
	LastName    string `json:"lastName"`              // The user's last name.
	Email       string `json:"email,omitempty"`       // The user's email address (optional for registration).
	MobilePhone string `json:"mobilePhone,omitempty"` // The user's mobile phone number (optional for registration).
	Login       string `json:"login"`                 // The user's login identifier (usually the same as email).
}

// User represents a simplified Okta user.
type User struct {
	ID      string `json:"id"`     // The user's unique ID in Okta.
	Status  string `json:"status"` // The user's status (e.g., "ACTIVE", "PROVISIONED").
	Profile struct {
		Email       string `json:"email"`       // The user's email address.
		MobilePhone string `json:"mobilePhone"` // The user's mobile phone number.
	} `json:"profile"` // The user's profile information.
}

// ErrorResponse represents a standard Okta API error response.
type ErrorResponse struct {
	ErrorCode    string `json:"errorCode"`    // The Okta error code.
	ErrorSummary string `json:"errorSummary"` // A summary of the error.
	ErrorLink    string `json:"errorLink"`    // A link to more information about the error.
	ErrorID      string `json:"errorId"`      // The unique ID of the error.
	ErrorCauses  []struct {
		ErrorSummary string `json:"errorSummary"` // The cause of the error.
	} `json:"errorCauses"` // An array of error causes.
}

// RegisterUser registers a new user with Okta.
// It supports registration with email, phone, or both.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The registration request data.
//
// Returns:
//   - The client ID upon successful registration.
//   - An error if the registration fails.
func (o *Auth) RegisterUser(ctx context.Context, req RegistrationRequest) (string, error) {
	// Validate that at least one of email or mobilePhone is provided.
	if req.Profile.Email == "" && req.Profile.MobilePhone == "" {
		return "", errors.New("at least one of email or mobilePhone must be provided for registration")
	}

	// If login is not provided, set it to email (if available) or mobilePhone.
	if req.Profile.Login == "" {
		if req.Profile.Email != "" {
			req.Profile.Login = req.Profile.Email
		} else {
			req.Profile.Login = req.Profile.MobilePhone // Consider using a more robust unique identifier if only phone is used.
		}
	}

	// Construct the API URL.
	url := fmt.Sprintf("%s/api/v1/users?activate=%t", o.Domain, req.Activate)

	// Marshal the request body to JSON.
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal registration request: %w", err)
	}

	// Make the API request.
	resp, err := o.makeRequest(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for successful status codes (200-299).
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		var user User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", fmt.Errorf("failed to decode user response: %w", err)
		}
		// Return the client ID on success.
		return o.ClientID, nil
	}

	// Handle API errors.
	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return "", fmt.Errorf("failed to decode error response (status: %d): %w", resp.StatusCode, err)
	}
	return "", fmt.Errorf("failed to register user (status: %d): %s", resp.StatusCode, errorResp.ErrorSummary)
}

// VerifyFactorRequest represents the request for factor verification
type VerifyFactorRequest struct {
	PassCode   string `json:"passCode"`   // The one-time passcode entered by the user.
	StateToken string `json:"stateToken"` // The state token received during factor challenge.
}

// VerifyFactorResponse represents response of verify factor
type VerifyFactorResponse struct {
	Status       string `json:"status"`       // Status of factor verification
	SessionToken string `json:"sessionToken"` // session token created after factor verification
}

// GetUser gets a user by ID or login (email or phone).
//
// Parameters:
//   - ctx: The context for the request.
//   - identifier: The user's ID or login.
//
// Returns:
//   - The user if found.
//   - An error if the user is not found or if an error occurs.
func (o *Auth) GetUser(ctx context.Context, identifier string) (*User, error) {
	// Construct the API URL.
	url := fmt.Sprintf("%s/api/v1/users/%s", o.Domain, identifier)

	// Make the API request.
	resp, err := o.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for successful status code (200 OK).
	if resp.StatusCode == http.StatusOK {
		var user User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, fmt.Errorf("failed to decode user response: %w", err)
		}
		return &user, nil
	}

	// Handle API errors.
	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return nil, fmt.Errorf("failed to decode error response (status: %d): %w", resp.StatusCode, err)
	}
	return nil, fmt.Errorf("failed to get user (status: %d): %s", resp.StatusCode, errorResp.ErrorSummary)
}

// GetUserFactors gets the factors enrolled for a user.
//
// Parameters:
//   - ctx: The context for the request.
//   - userID: The ID of the user.
//
// Returns:
//   - A list of factors.
//   - An error if the request fails.
func (o *Auth) GetUserFactors(ctx context.Context, userID string) ([]interface{}, error) {
	// Construct the API URL.
	url := fmt.Sprintf("%s/api/v1/users/%s/factors", o.Domain, userID)

	// Make the API request.
	resp, err := o.makeRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for successful status code (200 OK).
	if resp.StatusCode == http.StatusOK {
		var factors []interface{} // You can define a more specific struct if needed.
		if err := json.NewDecoder(resp.Body).Decode(&factors); err != nil {
			return nil, fmt.Errorf("failed to decode factors response: %w", err)
		}
		return factors, nil
	}

	// Handle API errors.
	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return nil, fmt.Errorf("failed to decode error response (status: %d): %w", resp.StatusCode, err)
	}
	return nil, fmt.Errorf("failed to get user factors (status: %d): %s", resp.StatusCode, errorResp.ErrorSummary)
}

// VerifyEmailOrPhone initiates the verification process for a user's email or phone.
// It sends a verification challenge (e.g., sends a one-time passcode).
//
// Parameters:
//   - ctx: The context for the request.
//   - identifier: The user's ID, email, or phone number.
//
// Returns:
//   - session token, which is needed to complete the verification.
//   - An error if the verification challenge cannot be initiated.
func (o *Auth) VerifyEmailOrPhone(ctx context.Context, identifier string) (string, error) {
	// 1. Get the user.
	user, err := o.GetUser(ctx, identifier)
	if err != nil {
		return "", err
	}

	// 2. Find the email or SMS factor.
	factors, err := o.GetUserFactors(ctx, user.ID)
	if err != nil {
		return "", err
	}

	var factorID string
	var factorType string

	// Iterate through the factors to find the appropriate email or SMS factor.
	for _, f := range factors {
		factorMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if the factor belongs to the Okta provider.
		provider, ok := factorMap["provider"].(string)
		if !ok || provider != "OKTA" {
			continue
		}

		// Get the factor type.
		fType, ok := factorMap["factorType"].(string)
		if !ok {
			continue
		}

		// Check if the factor type and identifier match.
		if (fType == "email" && user.Profile.Email != "" && (identifier == user.Profile.Email || identifier == user.ID)) ||
			(fType == "sms" && user.Profile.MobilePhone != "" && (identifier == user.Profile.MobilePhone || identifier == user.ID)) {
			factorID, _ = factorMap["id"].(string)
			factorType = fType
			break
		}
	}

	// Check if a suitable factor was found.
	if factorID == "" {
		return "", fmt.Errorf("email or SMS factor not found for user")
	}

	// 3. Trigger verification challenge.
	challengeURL := fmt.Sprintf("%s/api/v1/users/%s/factors/%s/verify", o.Domain, user.ID, factorID)

	// Make the API request to initiate the challenge.
	resp, err := o.makeRequest(ctx, http.MethodPost, challengeURL, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for successful status codes.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return "", fmt.Errorf("failed to decode error response (status: %d): %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("failed to trigger %s verification (status: %d): %s", factorType, resp.StatusCode, errorResp.ErrorSummary)
	}

	// Decode the verification response to extract the state token.
	var verifyResponse VerifyFactorResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResponse); err != nil {
		return "", fmt.Errorf("failed to decode verification response: %w", err)
	}

	// Return the session token.
	return verifyResponse.SessionToken, nil
}

// VerifyPasscode verifies the one-time passcode sent to the user's email or phone.
//
// Parameters:
//   - ctx: The context for the request.
//   - identifier: The user's ID, email, or phone number.
//   - passcode: The one-time passcode entered by the user.
//
// Returns:
//   - The client ID upon successful verification.
//   - An error if the verification fails.
func (o *Auth) VerifyPasscode(ctx context.Context, identifier string, passcode string) (string, error) {
	// 1. Get the user.
	user, err := o.GetUser(ctx, identifier)
	if err != nil {
		return "", err
	}

	// 2. Find the email or SMS factor.
	factors, err := o.GetUserFactors(ctx, user.ID)
	if err != nil {
		return "", err
	}

	var factorID string
	var factorType string

	// Iterate through the factors to find the appropriate email or SMS factor.
	for _, f := range factors {
		factorMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if the factor belongs to the Okta provider.
		provider, ok := factorMap["provider"].(string)
		if !ok || provider != "OKTA" {
			continue
		}

		// Get the factor type.
		fType, ok := factorMap["factorType"].(string)
		if !ok {
			continue
		}

		// Check if the factor type and identifier match.
		if (fType == "email" && user.Profile.Email != "" && (identifier == user.Profile.Email || identifier == user.ID)) ||
			(fType == "sms" && user.Profile.MobilePhone != "" && (identifier == user.Profile.MobilePhone || identifier == user.ID)) {
			factorID, _ = factorMap["id"].(string)
			factorType = fType
			break
		}
	}

	// Check if a suitable factor was found.
	if factorID == "" {
		return "", fmt.Errorf("email or SMS factor not found for user")
	}

	// 3. Verify the passcode.
	verifyURL := fmt.Sprintf("%s/api/v1/users/%s/factors/%s/verify", o.Domain, user.ID, factorID)

	// Create the verification request.
	verifyReq := VerifyFactorRequest{
		PassCode: passcode,
	}

	// Marshal the verification request to JSON.
	body, err := json.Marshal(verifyReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal verification request: %w", err)
	}

	// Make the API request to verify the passcode.
	resp, err := o.makeRequest(ctx, http.MethodPost, verifyURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for successful status codes (200-299).
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		// Return the client ID on successful verification.
		return o.ClientID, nil
	}

	// Handle API errors.
	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return "", fmt.Errorf("failed to decode error response (status: %d): %w", resp.StatusCode, err)
	}
	return "", fmt.Errorf("failed to verify %s (status: %d): %s", factorType, resp.StatusCode, errorResp.ErrorSummary)
}

// makeRequest is a helper function to make HTTP requests to the Okta API.
//
// Parameters:
//   - ctx: The context for the request.
//   - method: The HTTP method (e.g., "GET", "POST").
//   - urlStr: The URL for the request.
//   - body: The request body (can be nil for GET requests).
//
// Returns:
//   - The HTTP response.
//   - An error if the request fails.
func (o *Auth) makeRequest(ctx context.Context, method, urlStr string, body io.Reader) (*http.Response, error) {
	// Create a new HTTP request.
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required headers.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "SSWS "+o.APIToken)

	// Send the request using the Okta client's HTTP client.
	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		// Handle network errors, timeouts, etc.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if urlErr.Timeout() {
				return nil, fmt.Errorf("request timed out: %w", err)
			}
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
