import React, { useState } from "react"
import { OAuthConfig, OAuthConfigStatus, ImageUpload } from "@fider/models"
import { Failure, actions } from "@fider/services"
import { Form, Button, Input, SocialSignInButton, Field, ImageUploader, Toggle } from "@fider/components"
import { useFider } from "@fider/hooks"
import { HStack } from "@fider/components/layout"

interface OAuthFormProps {
  config?: OAuthConfig
  onCancel: () => void
  cantDisable: boolean
}

export const OAuthForm: React.FC<OAuthFormProps> = (props) => {
  const fider = useFider()
  const [provider] = useState((props.config && props.config.provider) || "")
  const [displayName, setDisplayName] = useState((props.config && props.config.displayName) || "")
  const [enabled, setEnabled] = useState((props.config && props.config.status === OAuthConfigStatus.Enabled) || false)
  const [clientID, setClientID] = useState((props.config && props.config.clientID) || "")
  const [clientSecret, setClientSecret] = useState((props.config && props.config.clientSecret) || "")
  const [clientSecretEnabled, setClientSecretEnabled] = useState(!props.config)
  const [authorizeURL, setAuthorizeURL] = useState((props.config && props.config.authorizeURL) || "")
  const [tokenURL, setTokenURL] = useState((props.config && props.config.tokenURL) || "")
  const [profileURL, setProfileURL] = useState((props.config && props.config.profileURL) || "")
  const [scope, setScope] = useState((props.config && props.config.scope) || "")
  const [jsonUserIDPath, setJSONUserIDPath] = useState((props.config && props.config.jsonUserIDPath) || "")
  const [jsonUserNamePath, setJSONUserNamePath] = useState((props.config && props.config.jsonUserNamePath) || "")
  const [jsonUserEmailPath, setJSONUserEmailPath] = useState((props.config && props.config.jsonUserEmailPath) || "")
  const [logo, setLogo] = useState<ImageUpload | undefined>()
  const [logoURL, setLogoURL] = useState<string | undefined>()
  const [logoBlobKey, setLogoBlobKey] = useState((props.config && props.config.logoBlobKey) || "")
  const [error, setError] = useState<Failure | undefined>()

  const handleSave = async () => {
    const result = await actions.saveOAuthConfig({
      provider,
      status: enabled ? OAuthConfigStatus.Enabled : OAuthConfigStatus.Disabled,
      displayName,
      clientID,
      clientSecret: clientSecretEnabled ? clientSecret : "",
      authorizeURL,
      tokenURL,
      profileURL,
      scope,
      jsonUserIDPath,
      jsonUserNamePath,
      jsonUserEmailPath,
      logo,
    })
    if (result.ok) {
      location.reload()
    } else {
      setError(result.error)
    }
  }

  const handleLogoChange = (newLogo: ImageUpload, instanceID: string, previewURL: string) => {
    setLogo(newLogo)
    setLogoURL(previewURL)
    setLogoBlobKey("")
  }

  const handleCancel = async () => {
    props.onCancel()
  }

  const enableClientSecret = () => {
    setClientSecret("")
    setClientSecretEnabled(true)
  }

  const title = props.config ? `OAuth Provider: ${props.config.displayName}` : "New OAuth Provider"
  return (
    <>
      <h2 className="text-title mb-2">{title}</h2>
      <Form error={error}>
        <div className="grid grid-cols-4 gap-4">
          <Input
            className="col-span-3"
            field="displayName"
            label="Display Name"
            maxLength={50}
            value={displayName}
            disabled={!fider.session.user.isAdministrator}
            onChange={setDisplayName}
          />
          <Field label="Button Preview">
            <SocialSignInButton option={{ displayName: displayName || "Button", provider, logoBlobKey, logoURL }} />
          </Field>
        </div>

        <ImageUploader label="Logo" field="logo" bkey={logoBlobKey} disabled={!fider.session.user.isAdministrator} onChange={handleLogoChange}>
          <p className="text-muted">
            We accept JPG, GIF and PNG images, smaller than 50KB and with an aspect ratio of 1:1 with minimum dimensions of 24x24 pixels.
          </p>
        </ImageUploader>

        <Input field="clientID" label="Client ID" maxLength={100} value={clientID} disabled={!fider.session.user.isAdministrator} onChange={setClientID} />

        <Input
          field="clientSecret"
          label="Client Secret"
          maxLength={500}
          value={clientSecret}
          disabled={!clientSecretEnabled}
          onChange={setClientSecret}
          afterLabel={
            !clientSecretEnabled ? (
              <>
                <span className="text-muted"> omitted for security reasons.</span>
                <span className="text-link text-normal text-xs ml-1" onClick={enableClientSecret}>
                  change
                </span>
              </>
            ) : undefined
          }
        />
        <Input
          field="authorizeURL"
          label="Authorize URL"
          maxLength={300}
          value={authorizeURL}
          disabled={!fider.session.user.isAdministrator}
          onChange={setAuthorizeURL}
        />
        <Input field="tokenURL" label="Token URL" maxLength={300} value={tokenURL} disabled={!fider.session.user.isAdministrator} onChange={setTokenURL} />

        <Input field="scope" label="Scope" maxLength={100} value={scope} disabled={!fider.session.user.isAdministrator} onChange={setScope}>
          <p className="text-muted">
            It is recommended to only request the minimum scopes we need to fetch the user <strong>id</strong>, <strong>name</strong> and <strong>email</strong>
            . Multiple scopes must be separated by space.
          </p>
        </Input>

        <h3 className="text-title mt-8 mb-2">User Profile</h3>
        <p className="text-muted">This section is used to configure how Fider will fetch user after the authentication process.</p>

        <Input
          field="profileURL"
          label="Profile API URL"
          maxLength={300}
          value={profileURL}
          disabled={!fider.session.user.isAdministrator}
          onChange={setProfileURL}
        >
          <p className="text-muted">The URL to fetch the authenticated user info. If empty, Fider will try to parse the user info from the Access Token.</p>
        </Input>

        <h3 className="text-title mt-8 mb-2">JSON Path</h3>

        <div className="grid grid-cols-3 gap-4">
          <Input
            field="jsonUserIDPath"
            label="ID"
            maxLength={100}
            value={jsonUserIDPath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserIDPath}
          >
            <p className="text-muted">
              Path to extract User ID from the JSON. This ID <strong>must</strong> be unique within the provider or unexpected side effects might happen. For
              example below, the path would be <strong>id</strong>.
            </p>
          </Input>
          <Input
            field="jsonUserNamePath"
            label="Name"
            maxLength={100}
            value={jsonUserNamePath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserNamePath}
          >
            <p className="text-muted">
              Path to extract user Display Name from the JSON. This is optional, but <strong>highly</strong> recommended. For the example below, the path would
              be <strong>profile.name</strong>.
            </p>
          </Input>
          <Input
            field="jsonUserEmailPath"
            label="Email"
            maxLength={100}
            value={jsonUserEmailPath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserEmailPath}
          >
            <p className="text-muted">
              Path to extract user Email from the JSON. This is optional, but <strong>highly</strong> recommended. For the example below, the path would be{" "}
              <strong>profile.emails[0]</strong>.
            </p>
          </Input>
        </div>

        <h3 className="text-title mb-2">Example Response</h3>

        <pre>
          {`{ 
  id: "35235"
  title: "Sr. Account Manager",
  profile: {
    dob: "01/05/2018",
    name: "John Doe"
    emails: [
      "john.doe@company.com"
    ]
  }
}
          `}
        </pre>

        {enabled && props.cantDisable && (
          <p className="text-red-700">You cannot disable this OAuth provider because this is your only one enabled and email auth is disabled</p>
        )}
        <Field label="Status">
          <Toggle field="status" disabled={props.cantDisable} active={enabled} onToggle={setEnabled} label={enabled ? "Enabled" : "Disabled"} />
          <div className="mt-1">
            {enabled ? (
              <p className="text-muted mt-1">
                This provider will be available for everyone to use during the sign in process. It is recommended that you keep it disable and test it before
                enabling it. The Test button is available after saving this configuration.
              </p>
            ) : (
              <p className="text-muted">Users won&apos;t be able to sign in with this Provider.</p>
            )}
          </div>
        </Field>

        <HStack className="mt-2">
          <Button variant="primary" onClick={handleSave}>
            Save
          </Button>
          <Button variant="tertiary" onClick={handleCancel}>
            Cancel
          </Button>
        </HStack>
      </Form>
    </>
  )
}
