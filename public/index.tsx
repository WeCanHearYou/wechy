import "@fider/assets/styles/index.scss"

import React, { Suspense } from "react"
import ReactDOM from "react-dom"
import { resolveRootComponent } from "@fider/router"
import { Header, ErrorBoundary, Loader } from "@fider/components"
import { classSet, Fider, FiderContext, actions } from "@fider/services"
import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import enJson from "./locales/en.json"
import jaJson from "./locales/ja.json"

const Loading = () => (
  <div className="page">
    <Loader />
  </div>
)

const logProductionError = (err: Error) => {
  if (Fider.isProduction()) {
    console.error(err)
    actions.logError(`react.ErrorBoundary: ${err.message}`, err)
  }
}

window.addEventListener("unhandledrejection", (evt: PromiseRejectionEvent) => {
  if (evt.reason instanceof Error) {
    actions.logError(`window.unhandledrejection: ${evt.reason.message}`, evt.reason)
  } else if (evt.reason) {
    actions.logError(`window.unhandledrejection: ${evt.reason.toString()}`)
  }
})

window.addEventListener("error", (evt: ErrorEvent) => {
  if (evt.error && evt.colno > 0 && evt.lineno > 0) {
    actions.logError(`window.error: ${evt.message}`, evt.error)
  }
})
;(() => {
  const fider = Fider.initialize()
  const defaultLanguage = "en"
  i18n.use(initReactI18next).init({
    debug: true,
    resources: {
      en: { translation: enJson },
      ja: { translation: jaJson },
    },
    lng: defaultLanguage,
    fallbackLng: false,
    returnEmptyString: false,
    keySeparator: ">",
    nsSeparator: "|",
  })

  __webpack_nonce__ = fider.session.contextID
  __webpack_public_path__ = `${fider.settings.assetsURL}/assets/`

  const config = resolveRootComponent(location.pathname)
  document.body.className = classSet({
    "is-authenticated": fider.session.isAuthenticated,
    "is-staff": fider.session.isAuthenticated && fider.session.user.isCollaborator,
  })
  ReactDOM.render(
    <React.StrictMode>
      <ErrorBoundary onError={logProductionError}>
        <FiderContext.Provider value={fider}>
          {config.showHeader && <Header />}
          <Suspense fallback={<Loading />}>{React.createElement(config.component, fider.session.props)}</Suspense>
        </FiderContext.Provider>
      </ErrorBoundary>
    </React.StrictMode>,
    document.getElementById("root")
  )
})()
