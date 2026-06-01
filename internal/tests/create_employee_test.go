package tests

import (
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "os"
  "testing"
  "time"
  "fmt"
  "math/rand"
  "sync"
)

func TestCreateEmployees(t *testing.T) {
  base := "http://localhost:8000"
  if v := os.Getenv("BASE_URL"); v != "" {
    base = v
  }

  accounts := 5
  employeesPerAccount := 3

  var wg sync.WaitGroup
  wg.Add(accounts)

  for a := 0; a < accounts; a++ {
    go func(a int) {
      defer wg.Done()

      jar, _ := cookiejar.New(nil)
      client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

      unique := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
      username := "user_" + unique
      password := "Pass!" + unique
      email := "user_" + unique + "@example.com"

      // register
      form := url.Values{}
      form.Set("username", username)
      form.Set("password", password)
      form.Set("email", email)

      resp, err := client.PostForm(base+"/register", form)
      if err != nil {
        t.Errorf("register error: %v", err)
        return
      }
      resp.Body.Close()

      // login
      loginForm := url.Values{}
      loginForm.Set("username", username)
      loginForm.Set("email", email)
      loginForm.Set("password", password)
      resp, err = client.PostForm(base+"/login", loginForm)
      if err != nil {
        t.Errorf("login error: %v", err)
        return
      }
      resp.Body.Close()

      // criar funcionários
      for i := 0; i < employeesPerAccount; i++ {
        ef := url.Values{}
        ef.Set("full_name", fmt.Sprintf("Emp %d %s", i, unique))
        ef.Set("email", fmt.Sprintf("emp_%d_%s@example.com", i, unique))
        ef.Set("position", "Engineer")

        resp, err := client.PostForm(base+"/employees", ef)
        if err != nil {
          t.Errorf("create employee error: %v", err)
          return
        }
        resp.Body.Close()

        time.Sleep(100 * time.Millisecond)
      }

    }(a)
  }

  wg.Wait()
}
