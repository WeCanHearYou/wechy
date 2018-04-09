import "./PrivacySettings.page.scss";

import * as React from "react";

import { SystemSettings, CurrentUser, Tenant } from "@fider/models";
import { Button, ButtonClickEvent, Textarea, DisplayError, Toggle } from "@fider/components/common";
import { actions, page, Failure } from "@fider/services";
import { AdminBasePage } from "../components";

interface PrivacySettingsPageProps {
  user: CurrentUser;
  tenant: Tenant;
}

interface PrivacySettingsPageState {
  isPrivate: boolean;
}

export class PrivacySettingsPage extends AdminBasePage<PrivacySettingsPageProps, PrivacySettingsPageState> {
  public name = "privacy";
  public icon = "key";
  public title = "Privacy";
  public subtitle = "Manage your site privacy";

  constructor(props: PrivacySettingsPageProps) {
    super(props);

    this.state = {
      isPrivate: this.props.tenant.isPrivate
    };

    page.setTitle(`Privacy · Site Settings · ${document.title}`);
  }

  private toggle = async (active: boolean) => {
    this.setState(
      state => ({
        isPrivate: active
      }),
      async () => {
        await actions.updateTenantPrivacy(this.state.isPrivate);
      }
    );
  };

  public content() {
    return (
      <div className="ui form">
        <div className="field">
          <label htmlFor="private">
            Private site
            <Toggle active={this.state.isPrivate} onToggle={this.toggle} />
          </label>
          <p className="info">
            A private site prevents unauthenticated users from viewing or interacting with its content. <br /> If
            enabled, only registered and invited users will be able to sign in to this site.
          </p>
        </div>
      </div>
    );
  }
}
