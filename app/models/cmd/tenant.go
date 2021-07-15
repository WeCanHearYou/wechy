package cmd

import (
	"time"

	"github.com/getfider/fider/app/models/dto"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
)

type CreateTenant struct {
	Name      string
	Subdomain string
	Status    int

	Result *entity.Tenant
}

type UpdateTenantPrivacySettings struct {
	IsPrivate bool
}

type UpdateTenantAllowingEmailAuthSettings struct {
	IsAllowingEmailAuth bool
}

type UpdateTenantSettings struct {
	Logo           *dto.ImageUpload
	Title          string
	Invitation     string
	WelcomeMessage string
	CNAME          string
	Locale          string
}

type UpdateTenantAdvancedSettings struct {
	CustomCSS string
}

type ActivateTenant struct {
	TenantID int
}

type SaveVerificationKey struct {
	Key      string
	Duration time.Duration
	Request  NewEmailVerification
}

//NewEmailVerification is used to define an email verification process
type NewEmailVerification interface {
	GetEmail() string
	GetName() string
	GetUser() *entity.User
	GetKind() enum.EmailVerificationKind
}

type SetKeyAsVerified struct {
	Key string
}
