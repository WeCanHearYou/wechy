package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"path"

	"github.com/getfider/fider/app/pkg/errors"
	"github.com/joeshaw/envdecode"
)

type config struct {
	Environment    string `env:"GO_ENV,default=production"`
	SignUpDisabled bool   `env:"SIGNUP_DISABLED,default=false"`
	AutoSSL        bool   `env:"SSL_AUTO,default=false"`
	SSLCert        string `env:"SSL_CERT"`
	SSLCertKey     string `env:"SSL_CERT_KEY"`
	Port           string `env:"PORT,default=3000"`
	HostMode       string `env:"HOST_MODE,default=single"`
	HostDomain     string `env:"HOST_DOMAIN,required"`
	JWTSecret      string `env:"JWT_SECRET,required"`
	Rendergun      struct {
		URL string `env:"RENDERGUN_URL"`
	}
	Database struct {
		URL          string `env:"DATABASE_URL,required"`
		MaxIdleConns int    `env:"DATABASE_MAX_IDLE_CONNS,default=2,strict"`
		MaxOpenConns int    `env:"DATABASE_MAX_OPEN_CONNS,default=4,strict"`
	}
	Stripe struct {
		SecretKey string `env:"STRIPE_SECRET_KEY"`
		PublicKey string `env:"STRIPE_PUBLIC_KEY"`
	}
	CDN struct {
		Host string `env:"CDN_HOST"`
	}
	Log struct {
		Level string `env:"LOG_LEVEL,default=INFO"`
	}
	OAuth struct {
		Google struct {
			ClientID string `env:"OAUTH_GOOGLE_CLIENTID"`
			Secret   string `env:"OAUTH_GOOGLE_SECRET"`
		}
		Facebook struct {
			AppID  string `env:"OAUTH_FACEBOOK_APPID"`
			Secret string `env:"OAUTH_FACEBOOK_SECRET"`
		}
		GitHub struct {
			ClientID string `env:"OAUTH_GITHUB_CLIENTID"`
			Secret   string `env:"OAUTH_GITHUB_SECRET"`
		}
	}
	Webhook struct {
		Slack struct {
			URL string `env:"WEBHOOK_SLACK_URL"`
		}
		Custom struct {
			URL string `env:"WEBHOOK_CUSTOM_URL"`
		}
	}
	Email struct {
		NoReply   string `env:"EMAIL_NOREPLY,required"`
		Whitelist string `env:"EMAIL_WHITELIST"`
		Blacklist string `env:"EMAIL_BLACKLIST"`
		Mailgun   struct {
			APIKey string `env:"EMAIL_MAILGUN_API"`
			Domain string `env:"EMAIL_MAILGUN_DOMAIN"`
			Region string `env:"EMAIL_MAILGUN_REGION,default=US"`
		}
		SMTP struct {
			Host     string `env:"EMAIL_SMTP_HOST"`
			Port     string `env:"EMAIL_SMTP_PORT"`
			Username string `env:"EMAIL_SMTP_USERNAME"`
			Password string `env:"EMAIL_SMTP_PASSWORD"`
		}
	}
	BlobStorage struct {
		Type string `env:"BLOB_STORAGE,default=sql"`
		S3   struct {
			EndpointURL     string `env:"BLOB_STORAGE_S3_ENDPOINT_URL"`
			Region          string `env:"BLOB_STORAGE_S3_REGION"`
			AccessKeyID     string `env:"BLOB_STORAGE_S3_ACCESS_KEY_ID"`
			SecretAccessKey string `env:"BLOB_STORAGE_S3_SECRET_ACCESS_KEY"`
			BucketName      string `env:"BLOB_STORAGE_S3_BUCKET"`
		}
		FS struct {
			Path string `env:"BLOB_STORAGE_FS_PATH"`
		}
	}
	Maintenance struct {
		Enabled bool   `env:"MAINTENANCE,default=false,strict"`
		Message string `env:"MAINTENANCE_MESSAGE"`
		Until   string `env:"MAINTENANCE_UNTIL"`
	}
	GoogleAnalytics string `env:"GOOGLE_ANALYTICS"`
}

// Config is a strongly typed reference to all configuration parsed from Environment Variables
var Config config

func init() {
	Reload()
}

// Reload configuration from current Enviornment Variables
func Reload() {
	Config = config{}
	err := envdecode.Decode(&Config)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse environment variables"))
	}

	if Config.Email.Mailgun.APIKey != "" {
		mustBeSet("EMAIL_MAILGUN_DOMAIN")
	} else {
		mustBeSet("EMAIL_SMTP_HOST")
		mustBeSet("EMAIL_SMTP_PORT")
	}

	bsType := strings.ToLower(Config.BlobStorage.Type)
	if bsType == "s3" {
		mustBeSet("BLOB_STORAGE_S3_BUCKET")
	} else if bsType == "fs" {
		mustBeSet("BLOB_STORAGE_FS_PATH")
	}
}

func mustBeSet(name string) {
	value := os.Getenv(name)
	if value == "" {
		panic(fmt.Errorf("Could not find environment variable named '%s'", name))
	}
}

// IsBillingEnabled returns true if billing is enabled
func IsBillingEnabled() bool {
	return Config.Stripe.SecretKey != ""
}

// IsSingleHostMode returns true if host mode is set to single tenant
func IsSingleHostMode() bool {
	return Config.HostMode == "single"
}

var hasLegal *bool

// HasLegal returns true if current instance contains legal documents: privacy.md and terms.md
func HasLegal() bool {
	if hasLegal == nil {
		_, err1 := os.Stat(Etc("privacy.md"))
		_, err2 := os.Stat(Etc("terms.md"))
		exists := err1 == nil && err2 == nil
		hasLegal = &exists
	}
	return *hasLegal
}

// MultiTenantDomain returns domain name of current instance for multi tenant hosts
func MultiTenantDomain() string {
	if !IsSingleHostMode() {
		return "." + Config.HostDomain
	}
	return ""
}

// IsProduction returns true on Fider production environment
func IsProduction() bool {
	return Config.Environment == "production" || (!IsTest() && !IsDevelopment())
}

// IsTest returns true on Fider test environment
func IsTest() bool {
	return Config.Environment == "test"
}

// IsDevelopment returns true on Fider production environment
func IsDevelopment() bool {
	return Config.Environment == "development"
}

// Path returns root path of project + given path
func Path(p ...string) string {
	root := "./"
	if IsTest() {
		_, b, _, _ := runtime.Caller(0)
		basepath := filepath.Dir(b)
		root = path.Join(basepath, "../../../")
	}

	elems := append([]string{root}, p...)
	return path.Join(elems...)
}

// Etc returns a path to a folder or file inside the /etc/ folder
func Etc(p ...string) string {
	paths := append([]string{"etc"}, p...)
	return Path(paths...)
}

// Subdomain returns the Fider subdomain (if available) from given host
func Subdomain(host string) string {
	if IsSingleHostMode() {
		return ""
	}

	domain := MultiTenantDomain()
	if domain != "" && strings.Contains(host, domain) {
		return strings.Replace(host, domain, "", -1)
	}

	if Config.CDN.Host != "" {
		domain = Config.CDN.Host
		parts := strings.Split(domain, ":")
		if parts[0] != "" && strings.Contains(host, "."+parts[0]) {
			return strings.Replace(host, "."+parts[0], "", -1)
		}
	}

	return ""
}
