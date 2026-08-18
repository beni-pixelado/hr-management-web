package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hr-management-web/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestRegisterCreatesOrganizationAndOwner(t *testing.T) {
	router := newRouter()
	client := newClient()

	form := url.Values{
		"username": []string{"auth_owner"},
		"password": []string{"Secret123!"},
		"email":    {"auth_owner@example.com"},
	}
	resp := post(t, client, "/register", form)
	defer resp.Body.Close()

	var user struct {
		Role           string
		OrganizationID uint
	}
	if err := testDB.Raw("SELECT role, organization_id FROM users WHERE username = ?", "auth_owner").Scan(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Role != "owner" {
		t.Errorf("expected owner role, got %q", user.Role)
	}
	if user.OrganizationID == 0 {
		t.Error("expected user to be assigned an organization")
	}

	_ = router
}

func TestLoginWrongPasswordFails(t *testing.T) {
	registerAndLogin(t, "wrongpw")

	client := newClient()
	form := url.Values{
		"username": []string{"wrongpw_owner"},
		"email":    {"wrongpw_owner@example.com"},
		"password": []string{"WrongPassword"},
	}
	resp := post(t, client, "/login", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad password, got %d", resp.StatusCode)
	}
}

func TestLogoutDestroysSession(t *testing.T) {
	client, _ := registerAndLogin(t, "logout")

	// logout is a GET route
	resp := clientGet(t, client, "/logout")
	resp.Body.Close()

	// After logout, a protected route must redirect to login.
	resp = clientGet(t, client, "/dashboard")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "/login" {
		t.Errorf("expected redirect to /login after logout, got %d -> %s", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestUnauthenticatedRedirectsToLogin(t *testing.T) {
	client := newClient()
	resp := clientGet(t, client, "/dashboard")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "/login" {
		t.Errorf("expected redirect to /login, got %d -> %s", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestRateLimitBlocksExcessRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(middleware.RateLimit(2, time.Minute))

	hits := 0
	r.GET("/limit", func(c *gin.Context) { hits++; c.Status(http.StatusOK) })

	rr := httptest.NewRecorder()
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/limit", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
	if hits != 2 {
		t.Errorf("expected handler to run only 2 times, ran %d", hits)
	}

	// A fresh recorder for the (blocked) third request must be 429.
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/limit", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 on 3rd request with limit 2, got %d", rr.Code)
	}
}

func clientGet(t *testing.T, client *http.Client, path string) *http.Response {
	t.Helper()
	resp, err := client.Get(serverURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}