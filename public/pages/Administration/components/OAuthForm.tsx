import React, { useState } from "react";
import { OAuthConfig, OAuthConfigStatus, ImageUpload } from "@fider/models";
import { Failure, actions } from "@fider/services";
import { Form, Button, Input, Heading, SocialSignInButton, Field, ImageUploader, Toggle } from "@fider/components";
import { useFider } from "@fider/hooks";

interface OAuthFormProps {
  config?: OAuthConfig;
  onCancel: () => void;
}

export const OAuthForm: React.FC<OAuthFormProps> = props => {
  const fider = useFider();
  const [provider] = useState((props.config && props.config.provider) || "");
  const [displayName, setDisplayName] = useState((props.config && props.config.displayName) || "");
  const [enabled, setEnabled] = useState((props.config && props.config.status === OAuthConfigStatus.Enabled) || false);
  const [isTrusted, setTrusted] = useState((props.config && props.config.isTrusted) || false);
  const [clientID, setClientID] = useState((props.config && props.config.clientID) || "");
  const [clientSecret, setClientSecret] = useState((props.config && props.config.clientSecret) || "");
  const [clientSecretEnabled, setClientSecretEnabled] = useState(!props.config);
  const [authorizeURL, setAuthorizeURL] = useState((props.config && props.config.authorizeURL) || "");
  const [tokenURL, setTokenURL] = useState((props.config && props.config.tokenURL) || "");
  const [profileURL, setProfileURL] = useState((props.config && props.config.profileURL) || "");
  const [scope, setScope] = useState((props.config && props.config.scope) || "");
  const [jsonUserIDPath, setJSONUserIDPath] = useState((props.config && props.config.jsonUserIDPath) || "");
  const [jsonUserNamePath, setJSONUserNamePath] = useState((props.config && props.config.jsonUserNamePath) || "");
  const [jsonUserEmailPath, setJSONUserEmailPath] = useState((props.config && props.config.jsonUserEmailPath) || "");
  const [logo, setLogo] = useState<ImageUpload | undefined>();
  const [logoURL, setLogoURL] = useState<string | undefined>();
  const [logoBlobKey, setLogoBlobKey] = useState((props.config && props.config.logoBlobKey) || "");
  const [error, setError] = useState<Failure | undefined>();

  const handleSave = async () => {
    const result = await actions.saveOAuthConfig({
      provider,
      status: enabled ? OAuthConfigStatus.Enabled : OAuthConfigStatus.Disabled,
      isTrusted,
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
      logo
    });
    if (result.ok) {
      location.reload();
    } else {
      setError(result.error);
    }
  };

  const handleLogoChange = (newLogo: ImageUpload, instanceID: string, previewURL: string) => {
    setLogo(newLogo);
    setLogoURL(previewURL);
    setLogoBlobKey("");
  };

  const handleCancel = async () => {
    props.onCancel();
  };

  const enableClientSecret = () => {
    setClientSecret("");
    setClientSecretEnabled(true);
  };

  const title = props.config ? `OAuth Provider: ${props.config.displayName}` : "New OAuth Provider";
  return (
    <>
      <Heading title={title} size="small" />
      <Form error={error}>
        <div className="row">
          <div className="col-sm-9">
            <Input
              field="displayName"
              label="Display Name"
              maxLength={50}
              value={displayName}
              disabled={!fider.session.user.isAdministrator}
              onChange={setDisplayName}
            />

            <ImageUploader
              label="Logo"
              field="logo"
              bkey={logoBlobKey}
              previewMaxWidth={80}
              disabled={!fider.session.user.isAdministrator}
              onChange={handleLogoChange}
            >
              <p className="info">
                We accept JPG, GIF and PNG images, smaller than 50KB and with an aspect ratio of 1:1 with minimum
                dimensions of 24x24 pixels.
              </p>
            </ImageUploader>
          </div>
          <div className="col-sm-3">
            <Field label="Button Preview">
              <SocialSignInButton option={{ displayName, provider, logoBlobKey, logoURL }} />
            </Field>
          </div>
        </div>

        <Input
          field="clientID"
          label="Client ID"
          maxLength={100}
          value={clientID}
          disabled={!fider.session.user.isAdministrator}
          onChange={setClientID}
        />

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
                <span className="info">omitted for security reasons.</span>
                <span className="info clickable" onClick={enableClientSecret}>
                  change
                </span>
              </>
            ) : (
              undefined
            )
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
        <Input
          field="tokenURL"
          label="Token URL"
          maxLength={300}
          value={tokenURL}
          disabled={!fider.session.user.isAdministrator}
          onChange={setTokenURL}
        />

        <Input
          field="scope"
          label="Scope"
          maxLength={100}
          value={scope}
          disabled={!fider.session.user.isAdministrator}
          onChange={setScope}
        >
          <p className="info">
            It is recommended to only request the minimum scopes we need to fetch the user <strong>id</strong>,{" "}
            <strong>name</strong> and <strong>email</strong>. Multiple scopes must be separated by space.
          </p>
        </Input>

        <h3>User Profile</h3>
        <p className="info">
          This section is used to configure how Fider will fetch user after the authentication process.
        </p>

        <Input
          field="profileURL"
          label="Profile API URL"
          maxLength={300}
          value={profileURL}
          disabled={!fider.session.user.isAdministrator}
          onChange={setProfileURL}
        >
          <p className="info">
            The URL to fetch the authenticated user info. If empty, Fider will try to parse the user info from the
            Access Token.
          </p>
        </Input>

        <h4>JSON Path</h4>

        <div className="row">
          <Input
            field="jsonUserIDPath"
            label="ID"
            className="col-sm-4"
            maxLength={100}
            value={jsonUserIDPath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserIDPath}
          >
            <p className="info">
              Path to extract User ID from the JSON. This ID <strong>must</strong> be unique within the provider or
              unexpected side effects might happen. For example below, the path would be <strong>id</strong>.
            </p>
          </Input>
          <Input
            field="jsonUserNamePath"
            label="Name"
            className="col-sm-4"
            maxLength={100}
            value={jsonUserNamePath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserNamePath}
          >
            <p className="info">
              Path to extract user Display Name from the JSON. This is optional, but <strong>highly</strong>{" "}
              recommended. For the example below, the path would be <strong>profile.name</strong>.
            </p>
          </Input>
          <Input
            field="jsonUserEmailPath"
            label="Email"
            className="col-sm-4"
            maxLength={100}
            value={jsonUserEmailPath}
            disabled={!fider.session.user.isAdministrator}
            onChange={setJSONUserEmailPath}
          >
            <p className="info">
              Path to extract user Email from the JSON. This is optional, but <strong>highly</strong> recommended. For
              the example below, the path would be <strong>profile.emails[0]</strong>.
            </p>
          </Input>
        </div>
        <pre>
          <h5>Example Response</h5>
          {`
  { 
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

        <div className="row">
          <div className="col-sm-4">
            <Field label="Trust this provider in private mode">
              <Toggle active={isTrusted} onToggle={setTrusted} />
              <span>{isTrusted ? "Trusted" : "Not Trusted"}</span>
              {isTrusted && (
                <p className="info">
                  In private mode, unknown users identified by this provider will be automatically created.
                </p>
              )}
              {!isTrusted && (
                <p className="info">
                  In private mode, users identified by this provider will be rejected 
                  if they are not already members or invited.
                </p>
              )}
            </Field>
          </div>
        </div>

        <div className="row">
          <div className="col-sm-4">
            <Field label="Status">
              <Toggle active={enabled} onToggle={setEnabled} />
              <span>{enabled ? "Enabled" : "Disabled"}</span>
              {enabled && (
                <p className="info">
                  This provider will be available for everyone to use during the sign in process. It is recommended that
                  you keep it disable and test it before enabling it. The Test button is available after saving this
                  configuration.
                </p>
              )}
              {!enabled && <p className="info">Users won't be able to sign in with this Provider.</p>}
            </Field>
          </div>
        </div>

        <div className="c-form-field">
          <Button color="positive" onClick={handleSave}>
            Save
          </Button>
          <Button color="cancel" onClick={handleCancel}>
            Cancel
          </Button>
        </div>
      </Form>
    </>
  );
};
