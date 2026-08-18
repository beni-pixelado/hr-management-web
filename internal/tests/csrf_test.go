package tests

import (
	"net/http"
	"net/url"
	"testing"
)

// A POST without a valid CSRF token must be rejected with 403.
func TestCSRFRejectsMissingToken(t *testing.T) {
	client, _ := registerAndLogin(t, "csrf_owner")

resp, err := client.PostForm(serverURL+"/employees", url.Values{
		"full_name": []string{"CSRF Victim"},
		"email":     []string{"csrf@example.com"},
		"position":  []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 without CSRF token, got %d", resp.StatusCode)
	}
}

// A POST with a bogus CSRF token must be rejected.
func TestCSRFRejectsWrongToken(t *testing.T) {
	client, _ := registerAndLogin(t, "csrf_wrong")

	resp, err := client.PostForm(serverURL+"/employees", url.Values{
		"_csrf":     []string{"not-the-real-token"},
		"full_name": []string{"CSRF Victim"},
		"email":     []string{"csrf2@example.com"},
		"position":  []string{"Engineer"},
	})
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 with wrong CSRF token, got %d", resp.StatusCode)
	}
}

// A POST carrying the token embedded in the rendered page succeeds.
func TestCSRFAcceptsValidToken(t *testing.T) {
	client, _ := registerAndLogin(t, "csrf_ok")

	resp := post(t, client, "/employees", url.Values{
		"full_name": []string{"CSRF Allowed"},
		"email":     {"csrf-ok@example.com"},
		"position":  {"Engineer"},
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302 with valid CSRF token, got %d", resp.StatusCode)
	}
}