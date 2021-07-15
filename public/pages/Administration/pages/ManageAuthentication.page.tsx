import React from "react"

import { Button, OAuthProviderLogo, Icon, Field, Toggle, Form } from "@fider/components"
import { OAuthConfig, OAuthProviderOption } from "@fider/models"
import { OAuthForm } from "../components/OAuthForm"
import { actions, notify, Fider, Failure } from "@fider/services"
import { AdminBasePage } from "../components/AdminBasePage"

import IconPlay from "@fider/assets/images/heroicons-play.svg"
import IconPencilAlt from "@fider/assets/images/heroicons-pencil-alt.svg"

import { HStack, VStack } from "@fider/components/layout"

interface ManageAuthenticationPageProps {
  providers: OAuthProviderOption[]
}

interface ManageAuthenticationPageState {
  isAdding: boolean
  isAllowingEmailAuth: boolean
  canDisableEmailAuth: boolean
  editing?: OAuthConfig
  error?: Failure
}

export default class ManageAuthenticationPage extends AdminBasePage<ManageAuthenticationPageProps, ManageAuthenticationPageState> {
  public id = "p-admin-authentication"
  public name = "authentication"
  public title = "Authentication"
  public subtitle = "Manage your site authentication"

  constructor(props: ManageAuthenticationPageProps) {
    super(props)
    this.state = {
      isAdding: false,
      isAllowingEmailAuth: Fider.session.tenant.isAllowingEmailAuth,
      canDisableEmailAuth: props.providers.map((o) => o.isEnabled).reduce((a, b) => a || b, false),
    }
  }

  private addNew = async () => {
    this.setState({ isAdding: true, editing: undefined })
  }

  private edit = async (provider: string) => {
    const result = await actions.getOAuthConfig(provider)
    if (result.ok) {
      this.setState({ editing: result.data, isAdding: false })
    } else {
      notify.error("Failed to retrieve OAuth configuration. Try again later")
    }
  }

  private startTest = async (provider: string) => {
    const redirect = `${Fider.settings.baseURL}/oauth/${provider}/echo`
    window.open(`/oauth/${provider}?redirect=${redirect}`, "oauth-test", "width=1100,height=600,status=no,menubar=no")
  }

  private cancel = async () => {
    this.setState({ isAdding: false, editing: undefined })
  }

  private toggleEmailAuth = async (active: boolean) => {
    this.setState(
      () => ({
        isAllowingEmailAuth: active,
      }),
      async () => {
        const response = await actions.updateTenantAllowingEmailAuth(this.state.isAllowingEmailAuth)
        if (response.ok) {
          notify.success(`You successfully ${this.state.isAllowingEmailAuth ? "allowed" : "disallowed"} email authentication.`)
        } else {
          this.setState(
            () => ({
              isAllowingEmailAuth: !active,
              error: response.error,
            }),
            () => notify.error("Unable to save this setting.")
          )
        }
      }
    )
  }

  public content() {
    let enabledProvidersCount = 0
    for (const o of this.props.providers) {
      if (o.isEnabled) {
        enabledProvidersCount++
      }
    }
    const cantDisable = !this.state.isAllowingEmailAuth && enabledProvidersCount == 1

    if (this.state.isAdding) {
      return <OAuthForm cantDisable={cantDisable} onCancel={this.cancel} />
    }

    if (this.state.editing) {
      return <OAuthForm cantDisable={cantDisable} config={this.state.editing} onCancel={this.cancel} />
    }

    const enabled = <span className="text-green-700">Enabled</span>
    const disabled = <span className="text-red-700">Disabled</span>

    return (
      <VStack spacing={8}>
        <div>
          <h2 className="text-display">General Authentication</h2>
          <Form error={this.state.error}>
            {this.state.isAllowingEmailAuth &&
              (this.state.canDisableEmailAuth ? (
                <p className="text-green-700">You can disable email auth because you have at least one authentication provider enabled</p>
              ) : (
                <p className="text-red-700">You cannot disable email auth because you do not have any authentication provider enabled</p>
              ))}
            <Field label="Allow Email Authentication">
              <Toggle
                field="isAllowingEmailAuth"
                disabled={!Fider.session.user.isAdministrator || !this.state.canDisableEmailAuth}
                active={this.state.isAllowingEmailAuth}
                onToggle={this.toggleEmailAuth}
              />
              <p className="text-muted my-1">
                When email-based authentication is disabled, users will not be allowed to sign in using their email. Thus, they will be forced to use another
                authentication provider, such as your prefered OAuth provider. <strong>Be sure to enable and test one before you turn this setting off!</strong>
              </p>
              <p className="text-muted mt-1">Note: Administrator accounts will still be allowed to sign in using their email.</p>
            </Field>
          </Form>
        </div>
        <div>
          <h2 className="text-display">OAuth Providers</h2>
          <p>
            You can use these section to add any authentication provider thats supports the OAuth2 protocol. Additional information is available in our{" "}
            <a rel="noopener" className="text-link" target="_blank" href="https://getfider.com/docs/configuring-oauth/">
              OAuth Documentation
            </a>
            .
          </p>
          <VStack spacing={6}>
            {this.props.providers.map((o) => (
              <div key={o.provider}>
                <HStack justify="between">
                  <HStack className="h-6">
                    <OAuthProviderLogo option={o} />
                    <strong>{o.displayName}</strong>
                  </HStack>
                  {o.isCustomProvider && (
                    <HStack>
                      {Fider.session.user.isAdministrator && (
                        <Button onClick={this.edit.bind(this, o.provider)} size="small">
                          <Icon sprite={IconPencilAlt} />
                          <span>Edit</span>
                        </Button>
                      )}
                      <Button onClick={this.startTest.bind(this, o.provider)} size="small">
                        <Icon sprite={IconPlay} />
                        <span>Test</span>
                      </Button>
                    </HStack>
                  )}
                </HStack>
                <div className="text-xs block my-1">{o.isEnabled ? enabled : disabled}</div>
                {o.isCustomProvider && (
                  <span className="text-muted">
                    <strong>Client ID:</strong> {o.clientID} <br />
                    <strong>Callback URL:</strong> {o.callbackURL}
                  </span>
                )}
              </div>
            ))}
            <div>
              {Fider.session.user.isAdministrator && (
                <Button variant="secondary" onClick={this.addNew}>
                  Add new
                </Button>
              )}
            </div>
          </VStack>
        </div>
      </VStack>
    )
  }
}
