package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig *oauth2.Config

	oauthStateString = "randomstring"

	globalUInfo *UserInfo
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback/login",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("SECRET_KEY"),
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}
}

type UserInfo struct {
	Name            string `json:"name"`
	AvatarUrl       string `json:"avatar_url"`
	Followers       int32  `json:"followers"`
	Following       int32  `json:"following"`
	Location        string `json:"location"`
	TwitterUsername string `json:"twitter_username"`
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	if globalUInfo == nil {
		http.Redirect(w, r, "/notlogged", http.StatusPermanentRedirect)
		return
	}

	var htmlIndex = fmt.Sprintf(`
		<html>
			<body>
				<img src="%v" />
				<div>%v</div>
				<div>Followers <span>%v</span> - Following <span>%v</span></div>
				<div>Location: %v</div>
				<div>Twitter: @%v</div>
			</body>
		</html>`, globalUInfo.AvatarUrl, globalUInfo.Name, globalUInfo.Followers, globalUInfo.Following, globalUInfo.Location, globalUInfo.TwitterUsername)

	fmt.Fprintf(w, htmlIndex)

}

func NoWherePage(w http.ResponseWriter, r *http.Request) {
	if globalUInfo != nil {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
	var htmlIndex = `
		<html>
			<body>
				<a href="/login">Github Log In</a>
			</body>
		</html>`

	fmt.Fprintf(w, htmlIndex)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("A?")
		http.Redirect(w, r, "/notlogged", http.StatusTemporaryRedirect)
		return
	}

	var userInfo *UserInfo
	err = json.Unmarshal(content, &userInfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	globalUInfo = userInfo

	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	return
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := githubOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+token.AccessToken)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}

func main() {

	http.HandleFunc("/", HomePage)
	http.HandleFunc("/notlogged", NoWherePage)
	http.HandleFunc("/login", LoginPage)
	http.HandleFunc("/callback/login", Callback)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
