import { http, Result } from "@fider/services/http";
import { UserRole, OAuthConfig, ImageUpload } from "@fider/models";

export interface CheckAvailabilityResponse {
  message: string;
}

export interface CreateTenantRequest {
  legalAgreement: boolean;
  tenantName: string;
  subdomain?: string;
  name?: string;
  token?: string;
  email?: string;
}

export interface CreateTenantResponse {
  token?: string;
}

export const createTenant = async (request: CreateTenantRequest): Promise<Result<CreateTenantResponse>> => {
  return await http.post<CreateTenantResponse>("/_api/tenants", request);
};

export interface UpdateTenantSettingsRequest {
  logo?: ImageUpload;
  title: string;
  invitation: string;
  welcomeMessage: string;
  cname: string;
}

export const updateTenantSettings = async (request: UpdateTenantSettingsRequest): Promise<Result> => {
  return await http.post("/_api/admin/settings/general", request);
};

export const updateTenantAdvancedSettings = async (customCSS: string): Promise<Result> => {
  return await http.post("/_api/admin/settings/advanced", { customCSS });
};

export const updateTenantPrivacy = async (isPrivate: boolean): Promise<Result> => {
  return await http.post("/_api/admin/settings/privacy", {
    isPrivate
  });
};

export const checkAvailability = async (subdomain: string): Promise<Result<CheckAvailabilityResponse>> => {
  return await http.get<CheckAvailabilityResponse>(`/_api/tenants/${subdomain}/availability`);
};

export const signIn = async (email: string): Promise<Result> => {
  return await http.post("/_api/signin", {
    email
  });
};

export const completeProfile = async (key: string, name: string): Promise<Result> => {
  return await http.post("/_api/signin/complete", {
    key,
    name
  });
};

export const changeUserRole = async (userID: number, role: UserRole): Promise<Result> => {
  return await http.post(`/_api/admin/roles/${role}/users`, {
    userID
  });
};

export const blockUser = async (userID: number): Promise<Result> => {
  return await http.put(`/_api/admin/users/${userID}/block`);
};

export const unblockUser = async (userID: number): Promise<Result> => {
  return await http.delete(`/_api/admin/users/${userID}/block`);
};

export const getOAuthConfig = async (provider: string): Promise<Result<OAuthConfig>> => {
  return await http.get<OAuthConfig>(`/_api/admin/oauth/${provider}`);
};

export interface CreateEditOAuthConfigRequest {
  provider: string;
  status: number;
  displayName: string;
  clientID: string;
  isTrusted: boolean;
  clientSecret: string;
  authorizeURL: string;
  tokenURL: string;
  scope: string;
  profileURL: string;
  jsonUserIDPath: string;
  jsonUserNamePath: string;
  jsonUserEmailPath: string;
  logo?: ImageUpload;
}

export const saveOAuthConfig = async (request: CreateEditOAuthConfigRequest): Promise<Result> => {
  return await http.post("/_api/admin/oauth", request);
};
