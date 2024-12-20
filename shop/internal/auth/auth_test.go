package auth

import (
	"context"
	"fmt"
	"log"
	"os"
)

func ExampleAuth_RegisterUser() {
	// Set your Auth configuration as environment variables
	AuthDomain := os.Getenv("Auth_DOMAIN")
	apiToken := os.Getenv("Auth_API_TOKEN")
	clientID := os.Getenv("Auth_CLIENT_ID")
	clientSecret := os.Getenv("Auth_CLIENT_SECRET")

	// Create a new Auth client
	AuthClient := New(AuthDomain, apiToken, clientID, clientSecret)

	// Create a context
	ctx := context.Background()

	// Prepare the registration request
	registrationRequest := RegistrationRequest{
		Profile: UserProfile{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Login:     "john.doe@example.com",
		},
		Activate:  true,
		SendEmail: true, // Send activation email if not activating immediately
	}

	// Register the user
	clientID, err := AuthClient.RegisterUser(ctx, registrationRequest)
	if err != nil {
		log.Fatalf("Failed to register user: %v", err)
	}

	fmt.Printf("User registered successfully. Client ID: %s\n", clientID)
}

func ExampleAuth_VerifyEmailOrPhone() {
	// Auth configuration (replace with your actual values)
	AuthDomain := os.Getenv("Auth_DOMAIN")
	apiToken := os.Getenv("Auth_API_TOKEN")
	clientID := os.Getenv("Auth_CLIENT_ID")
	clientSecret := os.Getenv("Auth_CLIENT_SECRET")

	// Create an Auth client.
	AuthClient := New(AuthDomain, apiToken, clientID, clientSecret)
	ctx := context.Background()

	// Assuming you have a user identifier (email or phone number)
	// Replace "user@example.com" with a real user identifier from your system
	userIdentifier := "john.doe@example.com"

	// Verify the user's email.
	sessionToken, err := AuthClient.VerifyEmailOrPhone(ctx, userIdentifier)
	if err != nil {
		log.Fatalf("Failed to initiate email/phone verification: %v", err)
	}

	fmt.Printf("Verification challenge initiated. sessionToken: %s\n", sessionToken)
	fmt.Println("Ask the user to enter the one-time passcode sent to their email/phone.")

	// Get the one-time passcode from the user (e.g., via standard input).
	var passcode string
	fmt.Print("Enter passcode: ")
	fmt.Scanln(&passcode)

	// Verify the passcode.
	clientID, err = AuthClient.VerifyPasscode(ctx, userIdentifier, passcode)
	if err != nil {
		log.Fatalf("Failed to verify passcode: %v", err)
	}

	fmt.Printf("Email/phone verified successfully. Client ID: %s\n", clientID)
}
