package postgres

import (
	"context"
	"time"

	"github.com/getfider/fider/app/pkg/bus"

	"github.com/getfider/fider/app/models/cmd"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/models/query"

	"github.com/getfider/fider/app/pkg/dbx"
	"github.com/getfider/fider/app/pkg/env"
	"github.com/getfider/fider/app/pkg/errors"
)

type dbTenant struct {
	ID                  int    `db:"id"`
	Name                string `db:"name"`
	Subdomain           string `db:"subdomain"`
	CNAME               string `db:"cname"`
	Invitation          string `db:"invitation"`
	WelcomeMessage      string `db:"welcome_message"`
	Status              int    `db:"status"`
	Locale              string `db:"locale"`
	IsPrivate           bool   `db:"is_private"`
	LogoBlobKey         string `db:"logo_bkey"`
	CustomCSS           string `db:"custom_css"`
	IsAllowingEmailAuth bool   `db:"is_allowing_email_auth"`
}

func (t *dbTenant) toModel() *entity.Tenant {
	if t == nil {
		return nil
	}

	tenant := &entity.Tenant{
		ID:                  t.ID,
		Name:                t.Name,
		Subdomain:           t.Subdomain,
		CNAME:               t.CNAME,
		Invitation:          t.Invitation,
		WelcomeMessage:      t.WelcomeMessage,
		Status:              t.Status,
		Locale:              t.Locale,
		IsPrivate:           t.IsPrivate,
		LogoBlobKey:         t.LogoBlobKey,
		CustomCSS:           t.CustomCSS,
		IsAllowingEmailAuth: t.IsAllowingEmailAuth,
	}

	return tenant
}

type dbEmailVerification struct {
	ID         int                        `db:"id"`
	Name       string                     `db:"name"`
	Email      string                     `db:"email"`
	Key        string                     `db:"key"`
	Kind       enum.EmailVerificationKind `db:"kind"`
	UserID     dbx.NullInt                `db:"user_id"`
	CreatedAt  time.Time                  `db:"created_at"`
	ExpiresAt  time.Time                  `db:"expires_at"`
	VerifiedAt dbx.NullTime               `db:"verified_at"`
}

func (t *dbEmailVerification) toModel() *entity.EmailVerification {
	model := &entity.EmailVerification{
		Name:       t.Name,
		Email:      t.Email,
		Key:        t.Key,
		Kind:       t.Kind,
		CreatedAt:  t.CreatedAt,
		ExpiresAt:  t.ExpiresAt,
		VerifiedAt: nil,
	}

	if t.VerifiedAt.Valid {
		model.VerifiedAt = &t.VerifiedAt.Time
	}

	if t.UserID.Valid {
		model.UserID = int(t.UserID.Int64)
	}

	return model
}

func isCNAMEAvailable(ctx context.Context, q *query.IsCNAMEAvailable) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		tenantID := 0
		if tenant != nil {
			tenantID = tenant.ID
		}

		exists, err := trx.Exists("SELECT id FROM tenants WHERE cname = $1 AND id <> $2", q.CNAME, tenantID)
		if err != nil {
			q.Result = false
			return errors.Wrap(err, "failed to check if tenant exists with CNAME '%s'", q.CNAME)
		}
		q.Result = !exists
		return nil
	})
}

func isSubdomainAvailable(ctx context.Context, q *query.IsSubdomainAvailable) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		exists, err := trx.Exists("SELECT id FROM tenants WHERE subdomain = $1", q.Subdomain)
		if err != nil {
			q.Result = false
			return errors.Wrap(err, "failed to check if tenant exists with subdomain '%s'", q.Subdomain)
		}
		q.Result = !exists
		return nil
	})
}

func updateTenantPrivacySettings(ctx context.Context, c *cmd.UpdateTenantPrivacySettings) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		_, err := trx.Execute("UPDATE tenants SET is_private = $1 WHERE id = $2", c.IsPrivate, tenant.ID)
		if err != nil {
			return errors.Wrap(err, "failed update tenant privacy settings")
		}
		return nil
	})
}

func updateTenantAllowingEmailAuthSettings(ctx context.Context, c *cmd.UpdateTenantAllowingEmailAuthSettings) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		_, err := trx.Execute("UPDATE tenants SET is_allowing_email_auth = $1 WHERE id = $2", c.IsAllowingEmailAuth, tenant.ID)
		if err != nil {
			return errors.Wrap(err, "failed update tenant allowing email auth settings")
		}
		return nil
	})
}

func updateTenantSettings(ctx context.Context, c *cmd.UpdateTenantSettings) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		if c.Logo.Remove {
			c.Logo.BlobKey = ""
		}

		query := "UPDATE tenants SET name = $1, invitation = $2, welcome_message = $3, cname = $4, logo_bkey = $5, locale = $6 WHERE id = $7"
		_, err := trx.Execute(query, c.Title, c.Invitation, c.WelcomeMessage, c.CNAME, c.Logo.BlobKey, c.Locale, tenant.ID)
		if err != nil {
			return errors.Wrap(err, "failed update tenant settings")
		}

		tenant.Name = c.Title
		tenant.Invitation = c.Invitation
		tenant.CNAME = c.CNAME
		tenant.WelcomeMessage = c.WelcomeMessage

		return nil
	})
}

func updateTenantAdvancedSettings(ctx context.Context, c *cmd.UpdateTenantAdvancedSettings) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		query := "UPDATE tenants SET custom_css = $1 WHERE id = $2"
		_, err := trx.Execute(query, c.CustomCSS, tenant.ID)
		if err != nil {
			return errors.Wrap(err, "failed update tenant advanced settings")
		}

		tenant.CustomCSS = c.CustomCSS
		return nil
	})
}

func activateTenant(ctx context.Context, c *cmd.ActivateTenant) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		query := "UPDATE tenants SET status = $1 WHERE id = $2"
		_, err := trx.Execute(query, enum.TenantActive, c.TenantID)
		if err != nil {
			return errors.Wrap(err, "failed to activate tenant with id '%d'", c.TenantID)
		}
		return nil
	})
}

func getVerificationByKey(ctx context.Context, q *query.GetVerificationByKey) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		verification := dbEmailVerification{}

		query := "SELECT id, email, name, key, created_at, verified_at, expires_at, kind, user_id FROM email_verifications WHERE key = $1 AND kind = $2 LIMIT 1"
		err := trx.Get(&verification, query, q.Key, q.Kind)
		if err != nil {
			return errors.Wrap(err, "failed to get email verification by its key")
		}

		q.Result = verification.toModel()
		return nil
	})
}

func saveVerificationKey(ctx context.Context, c *cmd.SaveVerificationKey) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		var userID interface{}
		if c.Request.GetUser() != nil {
			userID = c.Request.GetUser().ID
		}

		query := "INSERT INTO email_verifications (tenant_id, email, created_at, expires_at, key, name, kind, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
		_, err := trx.Execute(query, tenant.ID, c.Request.GetEmail(), time.Now(), time.Now().Add(c.Duration), c.Key, c.Request.GetName(), c.Request.GetKind(), userID)
		if err != nil {
			return errors.Wrap(err, "failed to save verification key for kind '%d'", c.Request.GetKind())
		}
		return nil
	})
}

func setKeyAsVerified(ctx context.Context, c *cmd.SetKeyAsVerified) error {
	return using(ctx, func(trx *dbx.Trx, tenant *entity.Tenant, user *entity.User) error {
		query := "UPDATE email_verifications SET verified_at = $1 WHERE tenant_id = $2 AND key = $3"
		_, err := trx.Execute(query, time.Now(), tenant.ID, c.Key)
		if err != nil {
			return errors.Wrap(err, "failed to update verified date of email verification request")
		}
		return nil
	})
}

func createTenant(ctx context.Context, c *cmd.CreateTenant) error {
	return using(ctx, func(trx *dbx.Trx, _ *entity.Tenant, _ *entity.User) error {
		now := time.Now()

		var id int
		err := trx.Get(&id,
			`INSERT INTO tenants (name, subdomain, created_at, cname, invitation, welcome_message, status, is_private, custom_css, logo_bkey, locale, is_allowing_email_auth) 
			 VALUES ($1, $2, $3, '', '', '', $4, false, '', '', $5, true) 
			 RETURNING id`, c.Name, c.Subdomain, now, c.Status, env.Config.Locale)
		if err != nil {
			return err
		}

		byDomain := &query.GetTenantByDomain{Domain: c.Subdomain}
		err = bus.Dispatch(ctx, byDomain)
		c.Result = byDomain.Result
		return err
	})
}

func getFirstTenant(ctx context.Context, q *query.GetFirstTenant) error {
	return using(ctx, func(trx *dbx.Trx, _ *entity.Tenant, _ *entity.User) error {
		tenant := dbTenant{}

		err := trx.Get(&tenant, `
			SELECT id, name, subdomain, cname, invitation, locale, welcome_message, status, is_private, logo_bkey, custom_css, is_allowing_email_auth
			FROM tenants
			ORDER BY id LIMIT 1
		`)

		if err != nil {
			return errors.Wrap(err, "failed to get first tenant")
		}

		q.Result = tenant.toModel()
		return nil
	})
}

func getTenantByDomain(ctx context.Context, q *query.GetTenantByDomain) error {
	return using(ctx, func(trx *dbx.Trx, _ *entity.Tenant, _ *entity.User) error {
		tenant := dbTenant{}

		err := trx.Get(&tenant, `
			SELECT id, name, subdomain, cname, invitation, locale, welcome_message, status, is_private, logo_bkey, custom_css, is_allowing_email_auth
			FROM tenants t
			WHERE subdomain = $1 OR subdomain = $2 OR cname = $3 
			ORDER BY cname DESC
		`, env.Subdomain(q.Domain), q.Domain, q.Domain)
		if err != nil {
			return errors.Wrap(err, "failed to get tenant with domain '%s'", q.Domain)
		}

		q.Result = tenant.toModel()
		return nil
	})
}
