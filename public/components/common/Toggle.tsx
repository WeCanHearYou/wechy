import "./Toggle.scss"

import React, { useState } from "react"
import { classSet } from "@fider/services"
import { HStack } from "../layout"
import { DisplayError, ValidationContext } from "@fider/components"

interface ToggleProps {
  field?: string
  label?: string
  active: boolean
  disabled?: boolean
  onToggle?: (active: boolean) => void
}

export const Toggle: React.FC<ToggleProps> = (props) => {
  const [active, setActive] = useState(props.active)

  const toggle = () => {
    if (props.disabled) {
      return
    }

    const newActive = !active
    setActive(newActive)
    if (props.onToggle) {
      props.onToggle(newActive)
    }
  }

  const className = classSet({
    "c-toggle": true,
    "c-toggle--enabled": active,
    "c-toggle--disabled": !!props.disabled,
  })

  return (
    <ValidationContext.Consumer>
      {(ctx) => (
        <>
          <HStack spacing={2}>
            <button onClick={toggle} type="button" className={className} role="switch">
              <span aria-hidden="true" className="shadow"></span>
            </button>
            {props.label && <span className="text-sm">{props.label}</span>}
          </HStack>
          {props.field && <DisplayError fields={[props.field]} error={ctx.error} />}
        </>
      )}
    </ValidationContext.Consumer>
  )
}
