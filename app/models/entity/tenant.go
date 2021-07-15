package entity

//Tenant represents a tenant
type Tenant struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Subdomain           string `json:"subdomain"`
	Invitation          string `json:"invitation"`
	WelcomeMessage      string `json:"welcomeMessage"`
	CNAME               string `json:"cname"`
	Status              int    `json:"status"`
	Locale              string `json:"locale"`
	IsPrivate           bool   `json:"isPrivate"`
	LogoBlobKey         string `json:"logoBlobKey"`
	CustomCSS           string `json:"-"`
	IsAllowingEmailAuth bool   `json:"isAllowingEmailAuth"`
}
