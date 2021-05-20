import React from "react"
import { Modal, Checkbox } from "@fider/components"
import { useFider } from "@fider/hooks"
import { useTranslation } from "react-i18next"

interface LegalAgreementProps {
  onChange: (agreed: boolean) => void
}

export const TermsOfService = () => {
  const fider = useFider()
  const { t } = useTranslation()

  if (fider.settings.hasLegal) {
    return (
      <a href="/terms" className="text-link" target="_blank">
        {t("Terms of Service")}
      </a>
    )
  }
  return null
}

export const PrivacyPolicy = () => {
  const fider = useFider()
  const { t } = useTranslation()

  if (fider.settings.hasLegal) {
    return (
      <a href="/privacy" className="text-link" target="_blank">
        {t("Privacy Policy")}
      </a>
    )
  }
  return null
}

export const LegalNotice = () => {
  const fider = useFider()
  const { t } = useTranslation()

  if (fider.settings.hasLegal) {
    return (
      <p className="text-muted">
        {t("By signing in, you agree to the ")}
        <PrivacyPolicy /> {t("and")} <TermsOfService />.
      </p>
    )
  }
  return null
}

export const LegalFooter = () => {
  const fider = useFider()

  if (fider.settings.hasLegal) {
    return (
      <Modal.Footer align="center">
        <LegalNotice />
      </Modal.Footer>
    )
  }
  return null
}

export const LegalAgreement: React.FunctionComponent<LegalAgreementProps> = (props) => {
  const fider = useFider()
  const { t } = useTranslation()

  if (fider.settings.hasLegal) {
    return (
      <Checkbox field="legalAgreement" onChange={props.onChange}>
        {t("I have read and agree to the ")}
        <PrivacyPolicy /> {t("and")} <TermsOfService />.
      </Checkbox>
    )
  }
  return null
}
