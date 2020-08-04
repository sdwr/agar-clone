package auth

import (
//    "log"

    "github.com/gorilla/mux"

    "github.com/qor/auth"
//    "github.com/qor/auth/auth_identity"
    "github.com/qor/auth/providers/github"
    "github.com/qor/auth/providers/google"
    "github.com/qor/auth/providers/password"
    "github.com/qor/auth/providers/facebook"
    "github.com/qor/auth/providers/twitter"
//    "github.com/qor/session/manager"

//    . "github.com/sdwr/agar-clone/types"
)

var (
    Auth = auth.New(&auth.Config{})
)

func initAuth() {
    Auth.RegisterProvider(password.New(&password.Config{}))
// Allow use Github
  Auth.RegisterProvider(github.New(&github.Config{
    ClientID:     "github client id",
    ClientSecret: "github client secret",
  }))

  // Allow use Google
  Auth.RegisterProvider(google.New(&google.Config{
    ClientID:     "google client id",
    ClientSecret: "google client secret",
    AllowedDomains: []string{}, // Accept all domains, instead you can pass a whitelist of acceptable domains
  }))

  // Allow use Facebook
  Auth.RegisterProvider(facebook.New(&facebook.Config{
    ClientID:     "facebook client id",
    ClientSecret: "facebook client secret",
  }))

  // Allow use Twitter
  Auth.RegisterProvider(twitter.New(&twitter.Config{
    ClientID:     "twitter client id",
    ClientSecret: "twitter client secret",
  }))
}

func LoadAuth(router *mux.Router ) {
    initAuth()
    router.HandleFunc("/auth/", Auth.NewServeMux().ServeHTTP)
}


