package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Username  string   `json:"username,omitempty"`
	Email     string   `json:"email,omitempty"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	Address   string   `json:"address,omitempty"`
	Address2  string   `json:"address2,omitempty"`
	City      string   `json:"city,omitempty"`
	State     string   `json:"state,omitempty"`
	Zip       string   `json:"zip,omitempty"`
	Country   string   `json:"country,omitempty"`
	Company   string   `json:"company,omitempty"`
	Website   string   `json:"website,omitempty"`
	Phone     string   `json:"phone,omitempty"`
	IsAdmin   bool     `json:"is_admin,omitempty"`
	IsSSO     bool     `json:"is_sso,omitempty"`
	UserType  string   `json:"user_type,omitempty"`
	Scopes    []string `json:"scopes,omitempty"`
}

// SubuserAccess represents a single subuser permission entry for a teammate.
type SubuserAccess struct {
	// ID is the numeric subuser account ID.
	ID int `json:"id"`
	// PermissionType is either "restricted" or "full".
	PermissionType string `json:"permission_type"`
	// Scopes is the list of scopes granted on this subuser (only meaningful for "restricted").
	Scopes []string `json:"scopes,omitempty"`
}

// SubuserAccessRead is the per-entry shape returned by GET /v3/teammates/{username}/subuser_access.
type SubuserAccessRead struct {
	ID             int      `json:"id"`
	Username       string   `json:"username,omitempty"`
	Email          string   `json:"email,omitempty"`
	Disabled       bool     `json:"disabled"`
	PermissionType string   `json:"permission_type"`
	Scopes         []string `json:"scopes,omitempty"`
}

// SubuserAccessResponse is the top-level body returned by GET /v3/teammates/{username}/subuser_access.
type SubuserAccessResponse struct {
	HasRestrictedSubuserAccess bool                `json:"has_restricted_subuser_access"`
	SubuserAccess              []SubuserAccessRead `json:"subuser_access"`
}

// The subuser access listing is paginated: it returns at most `limit` entries
// (100 when the parameter is omitted) and continues from `after_subuser_id`.
// Every caller here needs the whole list, because a partial list is what the
// next update would write back - the PATCH replaces subuser_access wholesale,
// so entries missing from a truncated read would lose access.
const (
	subuserAccessPageLimit = 500

	// Bounds the walk so a response that never advances the cursor cannot spin
	// forever. 500 * 40 is far beyond any real teammate.
	subuserAccessMaxPages = 40
)

// updateSSOTeammateRequest is the request body for PATCH /v3/sso/teammates/{username}.
// We keep it separate from User to avoid polluting that struct with SSO-only fields.
//
// is_admin and has_restricted_subuser_access both use omitempty so a zero value
// (false) is omitted rather than sent. This preserves the provider's existing
// "don't clobber out-of-band state" behaviour: an SSO teammate promoted to admin
// or granted restricted subuser access outside Terraform is not silently reset
// when an apply touches the teammate without managing those fields. The flag is
// only sent (true) when at least one subuser_access block is being managed.
type updateSSOTeammateRequest struct {
	FirstName                  string          `json:"first_name,omitempty"`
	LastName                   string          `json:"last_name,omitempty"`
	IsAdmin                    bool            `json:"is_admin,omitempty"`
	Scopes                     []string        `json:"scopes,omitempty"`
	HasRestrictedSubuserAccess bool            `json:"has_restricted_subuser_access,omitempty"`
	SubuserAccess              []SubuserAccess `json:"subuser_access,omitempty"`
}

type Users struct {
	Result []User `json:"result"`
}

type PendingUser struct {
	Result []struct {
		PendingID      string   `json:"pending_id,omitempty"`
		Token          string   `json:"token,omitempty"`
		Email          string   `json:"email,omitempty"`
		IsAdmin        bool     `json:"is_admin,omitempty"`
		IsReadOnly     bool     `json:"is_read_only,omitempty"`
		ExpirationDate int      `json:"expiration_date,omitempty"`
		Scopes         []string `json:"scopes,omitempty"`
	} `json:"result"`
}

func parseUser(respBody string) (*User, RequestError) {
	var body User

	err := json.Unmarshal([]byte(respBody), &body)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing teammate: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func (c *Client) GetUsernameByEmail(ctx context.Context, email string) (string, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/teammates?limit=10000")
	if err != nil {
		return "", RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	users := &Users{}

	decoder := json.NewDecoder(bytes.NewReader([]byte(respBody)))
	err = decoder.Decode(users)
	if err != nil {
		return "", RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	for _, user := range users.Result {
		if user.Email == email && user.Username != "" {
			return user.Username, RequestError{StatusCode: http.StatusOK, Err: nil}
		}
	}
	return "", RequestError{
		StatusCode: http.StatusNotFound,
		Err:        fmt.Errorf("username with email %s not found", email),
	}
}

func (c *Client) CreateUser(ctx context.Context, email string, scopes []string, isAdmin bool) (*User, RequestError) {
	respBody, statusCode, err := c.Post(ctx, "POST", "/teammates", User{
		Email:   email,
		IsAdmin: isAdmin,
		Scopes:  scopes,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseUser(respBody)
}

func (c *Client) CreateSSOUser(ctx context.Context, firstName, lastName, email string, scopes []string, isAdmin bool) (*User, RequestError) {
	respBody, statusCode, err := c.Post(ctx, "POST", "/sso/teammates", User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		IsAdmin:   isAdmin,
		Scopes:    scopes,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseUser(respBody)
}

func (c *Client) ReadUser(ctx context.Context, email string) (*User, RequestError) {
	username, requestErr := c.GetUsernameByEmail(ctx, email)
	if requestErr.Err != nil {
		// If user not found in active teammates, check pending invitations
		if requestErr.StatusCode == http.StatusNotFound {
			pendingUser, pendingErr := c.ReadPendingUser(ctx, email)
			if pendingErr.Err != nil {
				// User not found in either active or pending
				return nil, RequestError{
					StatusCode: http.StatusNotFound,
					Err:        fmt.Errorf("user with email %s not found in active teammates or pending invitations. Original active error: %v. Pending error: %v", email, requestErr.Err, pendingErr.Err),
				}
			}
			return pendingUser, RequestError{StatusCode: http.StatusOK, Err: nil}
		}
		return nil, requestErr
	}

	respBody, statusCode, err := c.Get(ctx, "GET", "/teammates/"+username)
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	var u User
	err = json.Unmarshal([]byte(respBody), &u)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}
	return &u, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func (c *Client) UpdateUser(ctx context.Context, email string, scopes []string, isAdmin bool) (*User, RequestError) {
	username, requestErr := c.GetUsernameByEmail(ctx, email)
	if requestErr.Err != nil {
		return nil, requestErr
	}

	respBody, statusCode, err := c.Post(ctx, "PATCH", "/teammates/"+username, User{
		IsAdmin: isAdmin,
		Scopes:  scopes,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseUser(respBody)
}

func (c *Client) UpdateSSOUser(ctx context.Context, firstName, lastName, email string, scopes []string, isAdmin bool) (*User, RequestError) {
	return c.UpdateSSOUserWithSubuserAccess(ctx, firstName, lastName, email, scopes, isAdmin, nil)
}

// UpdateSSOUserWithSubuserAccess calls PATCH /v3/sso/teammates/{username} and optionally sets
// subuser_access. When subuserAccess is nil or empty, has_restricted_subuser_access and
// subuser_access are omitted from the request body (see updateSSOTeammateRequest), so the
// teammate's existing subuser access is left untouched. Pass one or more entries to manage it;
// doing so sets has_restricted_subuser_access=true and replaces the access list.
func (c *Client) UpdateSSOUserWithSubuserAccess(
	ctx context.Context,
	firstName, lastName, email string,
	scopes []string,
	isAdmin bool,
	subuserAccess []SubuserAccess,
) (*User, RequestError) {
	username, requestErr := c.GetUsernameByEmail(ctx, email)
	if requestErr.Err != nil {
		if requestErr.StatusCode == http.StatusNotFound {
			return nil, RequestError{
				StatusCode: http.StatusNotFound,
				Err:        fmt.Errorf("user %s not found in active teammates - they may be pending and cannot be updated", email),
			}
		}
		return nil, requestErr
	}

	req := updateSSOTeammateRequest{
		FirstName:                  firstName,
		LastName:                   lastName,
		IsAdmin:                    isAdmin,
		Scopes:                     scopes,
		HasRestrictedSubuserAccess: len(subuserAccess) > 0,
		SubuserAccess:              subuserAccess,
	}

	respBody, statusCode, err := c.Post(ctx, "PATCH", "/sso/teammates/"+username, req)
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseUser(respBody)
}

// ReadSubuserAccess calls GET /v3/teammates/{username}/subuser_access and returns the current
// subuser access configuration for the given teammate email.
func (c *Client) ReadSubuserAccess(ctx context.Context, email string) (*SubuserAccessResponse, RequestError) {
	username, requestErr := c.GetUsernameByEmail(ctx, email)
	if requestErr.Err != nil {
		return nil, requestErr
	}

	// Walk the pages by cursor rather than trusting a page to be short: the API
	// may serve fewer entries than the requested limit, so "shorter than asked
	// for" does not mean "last page". Stop on an empty page, or as soon as a page
	// fails to advance past the cursor, which is what a server that ignores
	// after_subuser_id looks like.
	out := SubuserAccessResponse{}
	after := 0

	for page := 0; page < subuserAccessMaxPages; page++ {
		endpoint := fmt.Sprintf("/teammates/%s/subuser_access?limit=%d", username, subuserAccessPageLimit)
		if after > 0 {
			endpoint += fmt.Sprintf("&after_subuser_id=%d", after)
		}

		respBody, statusCode, err := c.Get(ctx, "GET", endpoint)
		if err != nil {
			return nil, RequestError{
				StatusCode: statusCode,
				Err:        fmt.Errorf("failed reading subuser_access for %s: %w", email, err),
			}
		}

		var body SubuserAccessResponse
		if jsonErr := json.Unmarshal([]byte(respBody), &body); jsonErr != nil {
			return nil, RequestError{
				StatusCode: http.StatusInternalServerError,
				Err:        fmt.Errorf("failed parsing subuser_access response: %w", jsonErr),
			}
		}

		if page == 0 {
			out.HasRestrictedSubuserAccess = body.HasRestrictedSubuserAccess
		}

		if len(body.SubuserAccess) == 0 {
			break
		}

		maxID := after
		for _, entry := range body.SubuserAccess {
			// Skip anything at or before the cursor: a server that ignores
			// after_subuser_id repeats the first page, and repeats must not
			// become duplicate entries.
			if entry.ID <= after {
				continue
			}
			out.SubuserAccess = append(out.SubuserAccess, entry)
			if entry.ID > maxID {
				maxID = entry.ID
			}
		}

		if maxID <= after {
			break
		}
		after = maxID
	}

	return &out, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func (c *Client) DeleteUser(ctx context.Context, email string) (bool, RequestError) {
	username, requestErr := c.GetUsernameByEmail(ctx, email)
	if requestErr.Err != nil {
		tokenInvite, tokenErr := c.GetPendingUserToken(ctx, email)
		if tokenErr.Err != nil {
			return false, tokenErr
		}

		if _, statusCode, err := c.Get(ctx, "DELETE", "/teammates/pending/"+tokenInvite); statusCode > 299 || err != nil {
			return false, RequestError{
				StatusCode: statusCode,
				Err:        fmt.Errorf("failed deleting user: %w", err),
			}
		}
		return false, RequestError{StatusCode: http.StatusOK, Err: nil}
	}

	if _, statusCode, err := c.Get(ctx, "DELETE", "/teammates/"+username); statusCode > 299 || err != nil {
		return false, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("failed deleting user: %w", err),
		}
	}

	return true, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func (c *Client) GetPendingUserToken(ctx context.Context, email string) (string, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/teammates/pending?limit=200")
	if err != nil {
		return "", RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	pendingUsers := &PendingUser{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(respBody)))
	err = decoder.Decode(pendingUsers)
	if err != nil {
		return "", RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	for _, user := range pendingUsers.Result {
		if user.Email == email {
			// SendGrid API returns token field, not pending_id
			if user.Token != "" {
				return user.Token, RequestError{StatusCode: http.StatusOK, Err: nil}
			}
			// Fallback to pending_id if token is empty (though this seems unlikely based on API response)
			if user.PendingID != "" {
				return user.PendingID, RequestError{StatusCode: http.StatusOK, Err: nil}
			}
		}
	}
	return "", RequestError{
		StatusCode: http.StatusNotFound,
		Err:        fmt.Errorf("pending user with email %s not found", email),
	}
}

// ReadPendingUser reads a pending user invitation by email
func (c *Client) ReadPendingUser(ctx context.Context, email string) (*User, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/teammates/pending?limit=10000")
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("failed to get pending users: %w", err),
		}
	}

	pendingUsers := &PendingUser{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(respBody)))
	err = decoder.Decode(pendingUsers)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to decode pending users response: %w", err),
		}
	}

	// Debug: log all pending users with more details
	var pendingDetails []string
	for _, pendingUser := range pendingUsers.Result {
		detail := fmt.Sprintf("email=%s, pending_id=%s, token=%s, expiration=%d",
			pendingUser.Email, pendingUser.PendingID, pendingUser.Token, pendingUser.ExpirationDate)
		pendingDetails = append(pendingDetails, detail)

		if pendingUser.Email == email {
			// Convert pending user to User struct
			user := &User{
				Email:   pendingUser.Email,
				IsAdmin: pendingUser.IsAdmin,
				Scopes:  pendingUser.Scopes,
				// Mark as pending by setting a special user type
				UserType: "pending",
			}
			return user, RequestError{StatusCode: http.StatusOK, Err: nil}
		}
	}

	return nil, RequestError{
		StatusCode: http.StatusNotFound,
		Err:        fmt.Errorf("pending user with email %s not found. Available pending users: %v. This may mean the user has already accepted the invitation or the invitation has expired", email, pendingDetails),
	}
}
