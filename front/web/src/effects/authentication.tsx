import React, { createContext, useState, useEffect, FC } from "react"
import firebase from "firebase"
import { auth } from "../firebase"

interface User extends firebase.UserInfo {
  readonly idToken: string
}

interface StatusSignedIn {
  readonly type: "signed-in"
  readonly user: User
}
const newSignedInStatus = (user: User): StatusSignedIn => ({
  type: "signed-in",
  user,
})

const statusUnauthenticated = { type: "unauthenticated" } as const
type StatusUnauthenticated = typeof statusUnauthenticated

const statusInitialized = { type: "initialized" } as const
type StatusInitialized = typeof statusInitialized

export type AuthenticationStatus = StatusInitialized | StatusUnauthenticated | StatusSignedIn

export const isSignedIn = (status: AuthenticationStatus): status is StatusSignedIn => status.type === "signed-in"

export const isUnauthenticated = (status: AuthenticationStatus): status is StatusUnauthenticated =>
  status.type === "unauthenticated"

export const isInitialized = (status: AuthenticationStatus): status is StatusInitialized =>
  status.type === "initialized"

const AuthenticationContext = createContext<AuthenticationStatus>(statusInitialized)

export const useAuthentication = (): AuthenticationStatus => {
  const [status, setStatus] = useState<AuthenticationStatus>(statusInitialized)

  useEffect(() => {
    return auth().onAuthStateChanged(async (user) => {
      if (user === null) {
        setStatus(statusUnauthenticated)
        return
      }

      const idToken = await user.getIdToken()
      debugger; // eslint-disable-line
      const { displayName, email, photoURL, phoneNumber, providerId, uid } = user
      setStatus(
        newSignedInStatus({
          idToken,
          displayName,
          email,
          phoneNumber,
          photoURL,
          providerId,
          uid,
        })
      )
    })
  }, [])

  return status
}

export const DefaultAuthenticationProvider: FC = ({ children }) => (
  <AuthenticationContext.Provider value={useAuthentication()}>{children}</AuthenticationContext.Provider>
)

const authProvider = new firebase.auth.GithubAuthProvider()

export const signIn = async (): Promise<void> => {
  await auth().signInWithPopup(authProvider)
}

export const signOut = async (): Promise<void> => {
  await auth().signOut()
}
