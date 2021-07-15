package mock

import (
	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models/entity"
	"github.com/getfider/fider/app/models/enum"
	"github.com/getfider/fider/app/pkg/env"
)

// DemoTenant is a mocked tenant
var DemoTenant *entity.Tenant

// AvengersTenant is a mocked tenant
var AvengersTenant *entity.Tenant

// JonSnow is a mocked user (administrator)
var JonSnow *entity.User

// AryaStark is a mocked user (visitor)
var AryaStark *entity.User

func init() {
	seed()
}

// NewSingleTenantServer creates a new multitenant test server
func NewSingleTenantServer() *Server {
	server := createServer()
	env.Config.HostMode = "single"
	return server
}

// NewServer creates a new server for HTTP testing
func NewServer() *Server {
	seed()
	server := createServer()
	env.Config.HostMode = "multi"
	return server
}

// NewWorker creates a new worker for worker testing
func NewWorker() *Worker {
	seed()
	worker := createWorker()
	return worker
}

func seed() {
	DemoTenant = &entity.Tenant{
		ID:                  1,
		Name:                "Demonstration",
		Subdomain:           "demo",
		Status:              enum.TenantActive,
		IsAllowingEmailAuth: true,
	}
	AvengersTenant = &entity.Tenant{
		ID:        2,
		Name:      "Avengers",
		Subdomain: "avengers",
		Status:    enum.TenantActive,
		CNAME:     "feedback.theavengers.com",
	}

	JonSnow = &entity.User{
		ID:     1,
		Name:   "Jon Snow",
		Email:  "jon.snow@got.com",
		Tenant: DemoTenant,
		Status: enum.UserActive,
		Role:   enum.RoleAdministrator,
		Providers: []*entity.UserProvider{
			{UID: "FB1234", Name: app.FacebookProvider},
		},
	}

	AryaStark = &entity.User{
		ID:     2,
		Name:   "Arya Stark",
		Email:  "arya.stark@got.com",
		Tenant: DemoTenant,
		Status: enum.UserActive,
		Role:   enum.RoleVisitor,
	}
}
