import React, { useState } from "react"

import { UserSettings } from "@fider/models"
import { Toggle, Field } from "@fider/components"
import { useFider } from "@fider/hooks"
import { HStack, VStack } from "@fider/components/layout"
import { useTranslation } from "react-i18next"

interface NotificationSettingsProps {
  userSettings: UserSettings
  settingsChanged: (settings: UserSettings) => void
}

type Channel = number
const WebChannel: Channel = 1
const EmailChannel: Channel = 2

export const NotificationSettings = (props: NotificationSettingsProps) => {
  const fider = useFider()
  const [userSettings, setUserSettings] = useState(props.userSettings)
  const { t } = useTranslation()

  const isEnabled = (settingsKey: string, channel: Channel): boolean => {
    if (settingsKey in userSettings) {
      return (parseInt(userSettings[settingsKey], 10) & channel) > 0
    }
    return false
  }

  const toggle = async (settingsKey: string, channel: Channel) => {
    const nextSettings = {
      ...userSettings,
      [settingsKey]: (parseInt(userSettings[settingsKey], 10) ^ channel).toString(),
    }
    setUserSettings(nextSettings)
    props.settingsChanged(nextSettings)
  }

  const icon = (settingsKey: string, channel: Channel) => {
    const active = isEnabled(settingsKey, channel)
    const label = channel === WebChannel ? "Web" : "Email"
    const onToggle = () => toggle(settingsKey, channel)
    return <Toggle key={`${settingsKey}_${channel}`} active={active} label={label} onToggle={onToggle} />
  }

  const info = (settingsKey: string, aboutForVisitors: string, aboutForCollaborators: string) => {
    const about = fider.session.user.isCollaborator ? aboutForCollaborators : aboutForVisitors
    const webEnabled = isEnabled(settingsKey, WebChannel)
    const emailEnabled = isEnabled(settingsKey, EmailChannel)

    if (!webEnabled && !emailEnabled) {
      return (
        <p className="text-muted">
          You&apos;ll <strong>NOT</strong> receive any notification about this event.
        </p>
      )
    } else if (webEnabled && !emailEnabled) {
      return (
        <p className="text-muted">
          You&apos;ll receive <strong>web</strong> notifications about {about}.
        </p>
      )
    } else if (!webEnabled && emailEnabled) {
      return (
        <p className="text-muted">
          You&apos;ll receive <strong>email</strong> notifications about {about}.
        </p>
      )
    } else if (webEnabled && emailEnabled) {
      return (
        <p className="text-muted">
          You&apos;ll receive <strong>web</strong> and <strong>email</strong> notifications about {about}.
        </p>
      )
    }
    return null
  }

  return (
    <>
      <Field label={t("Notifications")}>
        <p className="text-muted">{t("Use following panel to choose which events you&apos;d like to receive notification")}</p>

        <div className="notifications-settings">
          <VStack spacing={4} divide={true} className="p-2 bg-gray-50 rounded">
            <div>
              <div className="mb-1">New Post</div>
              {info("event_notification_new_post", "new posts on this site", "new posts on this site")}
              <HStack spacing={6}>
                {icon("event_notification_new_post", WebChannel)}
                {icon("event_notification_new_post", EmailChannel)}
              </HStack>
            </div>
            <div>
              <div className="mb-1">Discussion</div>
              {info("event_notification_new_comment", "comments on posts you've subscribed to", "comments on all posts unless individually unsubscribed")}
              <HStack spacing={6}>
                {icon("event_notification_new_comment", WebChannel)}
                {icon("event_notification_new_comment", EmailChannel)}
              </HStack>
            </div>
            <div>
              <div className="mb-1">Status Changed</div>
              {info(
                "event_notification_change_status",
                "status change on posts you've subscribed to",
                "status change on all posts unless individually unsubscribed"
              )}
              <HStack spacing={6}>
                {icon("event_notification_change_status", WebChannel)}
                {icon("event_notification_change_status", EmailChannel)}
              </HStack>
            </div>
          </VStack>
        </div>
      </Field>
    </>
  )
}
