export interface Tenant {
  id: number
  name: string
  cname: string
  subdomain: string
  locale: string
  invitation: string
  welcomeMessage: string
  status: TenantStatus
  isPrivate: boolean
  logoBlobKey: string
  isAllowingEmailAuth: boolean
}

export enum TenantStatus {
  Active = 1,
  Pending = 2,
  Locked = 3,
}

export interface User {
  id: number
  name: string
  role: UserRole
  status: UserStatus
  avatarURL: string
}

export enum UserAvatarType {
  Letter = "letter",
  Gravatar = "gravatar",
  Custom = "custom",
}

export enum UserStatus {
  Active = "active",
  Deleted = "deleted",
  Blocked = "blocked",
}

export enum UserRole {
  Visitor = "visitor",
  Collaborator = "collaborator",
  Administrator = "administrator",
}

export const isCollaborator = (role: UserRole): boolean => {
  return role === UserRole.Collaborator || role === UserRole.Administrator
}

export interface CurrentUser {
  id: number
  name: string
  email: string
  avatarType: UserAvatarType
  avatarBlobKey: string
  avatarURL: string
  role: UserRole
  status: UserStatus
  isAdministrator: boolean
  isCollaborator: boolean
}
