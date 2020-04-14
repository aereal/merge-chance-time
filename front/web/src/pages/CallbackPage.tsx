import React, { FC, useEffect } from "react"
import { useSignIn } from "../effects/authentication"

export const CallbackPage: FC = () => {
  const signIn = useSignIn()
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get("accessToken")
    if (token !== null) {
      signIn({ token })
      window.location.href = "/"
    }
  }, [window.location.search])
  return <>Redirecting...</>
}
