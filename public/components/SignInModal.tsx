import React, { useState, useEffect } from "react"
import { Modal, SignInControl, LegalFooter } from "@fider/components"
import { Button } from "./common"
import { useTranslation } from "react-i18next"

interface SignInModalProps {
  isOpen: boolean
  onClose: () => void
}

export const SignInModal: React.StatelessComponent<SignInModalProps> = (props) => {
  const [confirmationAddress, setConfirmationAddress] = useState("")
  const { t } = useTranslation()

  useEffect(() => {
    if (confirmationAddress) {
      setTimeout(() => setConfirmationAddress(""), 5000)
    }
  }, [confirmationAddress])

  const onEmailSent = (email: string): void => {
    setConfirmationAddress(email)
  }

  const closeModal = () => {
    setConfirmationAddress("")
    props.onClose()
  }

  const content = confirmationAddress ? (
    <>
      <p>
        {t("We have just sent a confirmation link to ")}
        <b>{confirmationAddress}</b>. <br /> {t("Click the link and you’ll be signed in.")}
      </p>
      <p>
        <Button variant="tertiary" onClick={closeModal}>
          OK
        </Button>
      </p>
    </>
  ) : (
    <SignInControl useEmail={true} onEmailSent={onEmailSent} />
  )

  return (
    <Modal.Window isOpen={props.isOpen} onClose={closeModal}>
      <Modal.Header>Sign in to post and vote</Modal.Header>
      <Modal.Content>{content}</Modal.Content>
      <LegalFooter />
    </Modal.Window>
  )
}
