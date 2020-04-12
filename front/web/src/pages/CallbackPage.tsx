import React, { FC, useEffect } from "react"

export const CallbackPage: FC = () => {
  useEffect(() => {
    const params = window.location.search
    if (window.parent) {
      window.parent.postMessage(JSON.stringify({ params }), window.location.origin)
      window.close()
    }
  }, [])
  return <>Redirecting...</>
}
