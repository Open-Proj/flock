package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Noah-Huppert/squad-up/server/models"
)

// Exchange users Google Id Token for a Squad Up API token, essentially the "login" endpoint.
func ExchangeTokenHandler (r *http.Request) models.HTTPResponse {
	resp := models.HTTPResponse
	// Get id_token passed in request
	idToken := r.PostFormValue("id_token")
	if len(idToken) == 0 {
		http.Error(w, "`id_token` must be provided as a post parameter", http.StatusUnprocessableEntity)
		return
	}

	// Make request to token info Gapi. This lets Google take care of
	// verifying the id token. If the token is valid it also provides
	// us with some basic profile info
	res, err := http.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + idToken)
	if err != nil {
		fmt.Printf("Error sending HTTP request to verify id token: %s\n", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	// Read response body
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Printf("Error reading body of response to verify id token %s\n", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	// Struct to unmarshal json resp into. Not all fields are
	// represented, only fields we care about.
	type GAPIIdTokenResp struct {
		// Id Token fields parsed by Gapi (Id Token is a JWT).
		// Audience field - Should be our apps Gapi client id.
		Aud string `json:"aud"`
		// Subject field - Authenticating user Gapi user id.
		Sub string `json:"sub"`

		// Google api fields
		// User email
		Email string `json:"email"`
		// If user has verified their email with Google
		EmailVerified bool `json:"email_verified,string"`
		// Url of profile picture
		Picture string `json:"picture"`
		// First name
		GivenName string `json:"given_name"`
		// Last name
		FamilyName string `json:"family_name"`
		// Locale string
		Locale string `json:"locale"`
	}

	// Decode response into json
	var resp GAPIIdTokenResp

	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Printf("Error decoding json response: %s\n", err)
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	// Check
	// Check that aud is our client id
	if resp.Aud != models.GapiConf.ClientId {
		http.Error(w, "Invalid id token", http.StatusUnauthorized)
		return
	}

	// Check that email is verified
	if resp.EmailVerified == false {
		http.Error(w, "Email not verified", http.StatusUnauthorized)
		return
	}

}