import React from "react"
import { SignInControl, Modal, Button, DisplayError, Form, Input, Message, LegalAgreement } from "@fider/components"
import { jwt, actions, Failure, querystring, Fider } from "@fider/services"
import { withTranslation } from "react-i18next"

interface OAuthUser {
  token: string
  name: string
  email: string
}

interface SignUpPageState {
  submitted: boolean
  tenantName: string
  legalAgreement: boolean
  error?: Failure
  name?: string
  email?: string
  subdomain: {
    available: boolean
    message?: string
    value?: string
  }
}

class SignUpPage extends React.Component<any, SignUpPageState> {
  private user?: OAuthUser

  constructor(props: any) {
    super(props)
    this.state = {
      submitted: false,
      legalAgreement: false,
      tenantName: "",
      subdomain: { available: false },
    }

    const token = querystring.get("token")
    if (token) {
      const data = jwt.decode(token)
      if (data) {
        this.user = {
          token,
          name: data["oauth/name"],
          email: data["oauth/email"],
        }
      }
    }
  }

  private confirm = async () => {
    const result = await actions.createTenant({
      token: this.user && this.user.token,
      legalAgreement: this.state.legalAgreement,
      tenantName: this.state.tenantName,
      subdomain: this.state.subdomain.value,
      name: this.state.name,
      email: this.state.email,
    })

    if (result.ok) {
      if (this.user) {
        if (Fider.isSingleHostMode()) {
          location.reload()
        } else {
          let baseURL = `${location.protocol}//${this.state.subdomain.value}${Fider.settings.domain}`
          if (location.port) {
            baseURL = `${baseURL}:${location.port}`
          }

          location.href = baseURL
        }
      } else {
        this.setState({ submitted: true })
      }
    } else if (result.error) {
      this.setState({ error: result.error, submitted: false })
    }
  }

  private timer?: number
  private checkAvailability = (subdomain: string) => {
    window.clearTimeout(this.timer)
    this.timer = window.setTimeout(() => {
      actions.checkAvailability(subdomain).then((result) => {
        this.setState({
          subdomain: {
            value: subdomain,
            available: !result.data.message,
            message: result.data.message,
          },
        })
      })
    }, 500)
  }

  private setSubdomain = async (subdomain: string) => {
    this.setState(
      {
        subdomain: {
          value: subdomain,
          available: false,
        },
      },
      this.checkAvailability.bind(this, subdomain)
    )
  }

  private onAgree = (agreed: boolean): void => {
    this.setState({ legalAgreement: agreed })
  }

  private setName = (name: string): void => {
    this.setState({ name })
  }

  private setEmail = (email: string): void => {
    this.setState({ email })
  }

  private setTenantName = (tenantName: string): void => {
    this.setState({ tenantName })
  }

  private noop = () => {
    // do nothing
  }

  public render() {
    const { t } = this.props
    const modal = (
      <Modal.Window canClose={false} isOpen={this.state.submitted} onClose={this.noop}>
        <Modal.Header>{t("Thank you for registering!")}</Modal.Header>
        <Modal.Content>
          <p>
            {t("We have just sent a confirmation link to")} <b>{this.state.email}</b>.
          </p>
          <p>{t("Click the link to complete the registration.")}</p>
        </Modal.Content>
      </Modal.Window>
    )

    return (
      <div id="p-signup" className="page container w-max-6xl">
        {modal}
        <div className="h-20 text-center mb-4">
          <img className="logo" alt="Logo" src="https://getfider.com/images/logo-100x100.png" />
        </div>

        <h3 className="text-display mb-2">{t("1. Who are you?")}</h3>
        <DisplayError fields={["token"]} error={this.state.error} />

        {this.user ? (
          <p>
            Hello, <b>{this.user.name}</b> {this.user.email && `(${this.user.email})`}
          </p>
        ) : (
          <>
            <p>{t("We need to identify you to setup your new Fider account.")}</p>
            <SignInControl useEmail={false} />
            <Form error={this.state.error}>
              <Input field="name" maxLength={100} onChange={this.setName} placeholder={t("Name")} />
              <Input field="email" maxLength={200} onChange={this.setEmail} placeholder={t("Email")} />
            </Form>
          </>
        )}

        <h3 className="text-display mb-2 mt-8">{t("2. What is this Feedback Forum for?")}</h3>

        <Form error={this.state.error} className="mb-8">
          <Input field="tenantName" maxLength={60} onChange={this.setTenantName} placeholder={t("your company or product name")} />
          {!Fider.isSingleHostMode() && (
            <Input field="subdomain" maxLength={40} onChange={this.setSubdomain} placeholder="subdomain" suffix={Fider.settings.domain}>
              {this.state.subdomain.available && (
                <Message className="mt-2" type="success" showIcon={true}>
                  This subdomain is available!
                </Message>
              )}
              {this.state.subdomain.message && (
                <Message className="mt-2" type="error" showIcon={true}>
                  {this.state.subdomain.message}
                </Message>
              )}
            </Input>
          )}
        </Form>

        <Form error={this.state.error} className="mb-4">
          <LegalAgreement onChange={this.onAgree} />
        </Form>

        <Button variant="primary" size="large" onClick={this.confirm}>
          {t("Confirm")}
        </Button>
      </div>
    )
  }
}

export default withTranslation()(SignUpPage)
