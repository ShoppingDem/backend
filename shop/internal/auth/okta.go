package auth

import (
	"context"
	"errors"
	"os"

	"github.com/okta/okta-sdk-golang/v5/okta"
)

var client *okta.APIClient

func init() {
	config := okta.NewConfiguration()
	config.OktaDomain = os.Getenv("OKTA_ORG_URL")
	client = okta.NewAPIClient(config)
}

func RegisterUser(ctx context.Context, email, phoneNumber string) (string, error) {
	user := okta.User{
		Profile: &okta.UserProfile{
			Email:       email,
			MobilePhone: phoneNumber,
		},
	}

	req := client.UserAPI.CreateUser(ctx)
	req = req.Body(user)
	createdUser, _, err := req.Execute()
	if err != nil {
		return "", err
	}

	return createdUser.Id, nil
}

func ValidateUser(ctx context.Context, email, phoneNumber string) (string, error) {
	filter := ""
	if email != "" {
		filter = "profile.email eq \"" + email + "\""
	} else if phoneNumber != "" {
		filter = "profile.phoneNumber eq \"" + phoneNumber + "\""
	} else {
		return "", errors.New("either email or phone number must be provided")
	}

	req := client.UserAPI.ListUsers(ctx)
	req = req.Filter(filter)
	users, _, err := req.Execute()
	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", errors.New("user not found")
	}

	return users[0].Id, nil
}
