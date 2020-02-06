package postgres

import (
	"context"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models/cmd"

	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/models/query"
	"github.com/getfider/fider/app/pkg/dbx"
	"github.com/getfider/fider/app/pkg/errors"
)

type dbOAuthConfig struct {
	ID                int    `db:"id"`
	Provider          string `db:"provider"`
	DisplayName       string `db:"display_name"`
	LogoBlobKey       string `db:"logo_bkey"`
	Status            int    `db:"status"`
	IsTrusted         bool   `db:"is_trusted"`
	ClientID          string `db:"client_id"`
	ClientSecret      string `db:"client_secret"`
	AuthorizeURL      string `db:"authorize_url"`
	TokenURL          string `db:"token_url"`
	Scope             string `db:"scope"`
	ProfileURL        string `db:"profile_url"`
	JSONUserIDPath    string `db:"json_user_id_path"`
	JSONUserNamePath  string `db:"json_user_name_path"`
	JSONUserEmailPath string `db:"json_user_email_path"`
}

func (m *dbOAuthConfig) toModel() *models.OAuthConfig {
	return &models.OAuthConfig{
		ID:                m.ID,
		Provider:          m.Provider,
		DisplayName:       m.DisplayName,
		Status:            m.Status,
		IsTrusted:         m.IsTrusted,
		LogoBlobKey:       m.LogoBlobKey,
		ClientID:          m.ClientID,
		ClientSecret:      m.ClientSecret,
		AuthorizeURL:      m.AuthorizeURL,
		TokenURL:          m.TokenURL,
		ProfileURL:        m.ProfileURL,
		Scope:             m.Scope,
		JSONUserIDPath:    m.JSONUserIDPath,
		JSONUserNamePath:  m.JSONUserNamePath,
		JSONUserEmailPath: m.JSONUserEmailPath,
	}
}

func getCustomOAuthConfigByProvider(ctx context.Context, q *query.GetCustomOAuthConfigByProvider) error {
	return using(ctx, func(trx *dbx.Trx, tenant *models.Tenant, user *models.User) error {
		if tenant == nil {
			return app.ErrNotFound
		}

		config := &dbOAuthConfig{}
		err := trx.Get(config, `
		SELECT id, provider, display_name, status, is_trusted, logo_bkey,
					 client_id, client_secret, authorize_url,
					 profile_url, token_url, scope, json_user_id_path,
					 json_user_name_path, json_user_email_path
		FROM oauth_providers
		WHERE tenant_id = $1 AND provider = $2
		`, tenant.ID, q.Provider)
		if err != nil {
			return err
		}

		q.Result = config.toModel()
		return nil
	})
}

func listCustomOAuthConfig(ctx context.Context, q *query.ListCustomOAuthConfig) error {
	return using(ctx, func(trx *dbx.Trx, tenant *models.Tenant, user *models.User) error {
		configs := []*dbOAuthConfig{}
		if tenant != nil {
			err := trx.Select(&configs, `
			SELECT id, provider, display_name, status, is_trusted, logo_bkey,
						 client_id, client_secret, authorize_url,
						 profile_url, token_url, scope, json_user_id_path,
						 json_user_name_path, json_user_email_path
			FROM oauth_providers
			WHERE tenant_id = $1
			ORDER BY id`, tenant.ID)
			if err != nil {
				return err
			}
		}

		q.Result = make([]*models.OAuthConfig, len(configs))
		for i, config := range configs {
			q.Result[i] = config.toModel()
		}
		return nil
	})
}

func saveCustomOAuthConfig(ctx context.Context, c *cmd.SaveCustomOAuthConfig) error {
	return using(ctx, func(trx *dbx.Trx, tenant *models.Tenant, user *models.User) error {
		var err error

		if c.Config.Logo.Remove {
			c.Config.Logo.BlobKey = ""
		}

		if c.Config.ID == 0 {
			query := `INSERT INTO oauth_providers (
				tenant_id, provider, display_name, status, is_trusted,
				client_id, client_secret, authorize_url,
				profile_url, token_url, scope, json_user_id_path,
				json_user_name_path, json_user_email_path, logo_bkey
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING id`

			err = trx.Get(&c.Config.ID, query, tenant.ID, c.Config.Provider,
				c.Config.DisplayName, c.Config.Status, c.Config.IsTrusted, c.Config.ClientID, c.Config.ClientSecret,
				c.Config.AuthorizeURL, c.Config.ProfileURL, c.Config.TokenURL,
				c.Config.Scope, c.Config.JSONUserIDPath, c.Config.JSONUserNamePath,
				c.Config.JSONUserEmailPath, c.Config.Logo.BlobKey)
		} else {
			query := `
				UPDATE oauth_providers 
				SET display_name = $3, status = $4, client_id = $5, client_secret = $6, 
						authorize_url = $7, profile_url = $8, token_url = $9, scope = $10, 
						json_user_id_path = $11, json_user_name_path = $12, json_user_email_path = $13,
						logo_bkey = $14, is_trusted = $15
			WHERE tenant_id = $1 AND id = $2`

			_, err = trx.Execute(query, tenant.ID, c.Config.ID,
				c.Config.DisplayName, c.Config.Status, c.Config.ClientID, c.Config.ClientSecret,
				c.Config.AuthorizeURL, c.Config.ProfileURL, c.Config.TokenURL,
				c.Config.Scope, c.Config.JSONUserIDPath, c.Config.JSONUserNamePath,
				c.Config.JSONUserEmailPath, c.Config.Logo.BlobKey, c.Config.IsTrusted)
		}

		if err != nil {
			return errors.Wrap(err, "failed to save OAuth Provider")
		}

		return nil
	})
}
