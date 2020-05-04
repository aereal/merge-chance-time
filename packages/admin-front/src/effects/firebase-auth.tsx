import React, { createContext, useState, useContext, useEffect, FC } from "react"
import { auth } from "../firebase"

export interface User {
  readonly uid: string
  readonly providerId: string
  readonly displayName: string | null
  readonly email: string | null
  readonly photoURL: string | null
  readonly token: string
  readonly claims: { [key: string]: any }
}

interface StatusSignedIn {
  readonly type: "signed-in"
  readonly user: User
}
export const newSignedInStatus = (user: User): StatusSignedIn => ({
  type: "signed-in",
  user,
})

export const statusUnauthenticated = { type: "unauthenticated" } as const
type StatusUnauthenticated = typeof statusUnauthenticated

export const statusInitialized = { type: "initialized" } as const
type StatusInitialized = typeof statusInitialized

export type AuthenticationStatus = StatusInitialized | StatusUnauthenticated | StatusSignedIn

export const isSignedIn = (status: AuthenticationStatus): status is StatusSignedIn => status.type === "signed-in"

export const isUnauthenticated = (status: AuthenticationStatus): status is StatusUnauthenticated =>
  status.type === "unauthenticated"

export const isInitialized = (status: AuthenticationStatus): status is StatusInitialized =>
  status.type === "initialized"

const FirebaseAuthContext = createContext<AuthenticationStatus>(statusInitialized)

export const DefaultFirebaseAuthProvider: FC = ({ children }) => (
  <FirebaseAuthContext.Provider value={useFirebaseAuth()}>{children}</FirebaseAuthContext.Provider>
)

export const useFirebaseAuth = (): AuthenticationStatus => {
  const [firebaseAuthStatus, setFirebaseAuthStatus] = useState(useContext(FirebaseAuthContext))
  useEffect(() => {
    return auth().onAuthStateChanged(async (user) => {
      if (user === null) {
        setFirebaseAuthStatus(statusUnauthenticated)
        return
      }

      const { displayName, email, photoURL, providerId, uid } = user
      const { token, claims } = await user.getIdTokenResult()
      setFirebaseAuthStatus(
        newSignedInStatus({
          displayName,
          email,
          photoURL,
          providerId,
          uid,
          token,
          claims,
        })
      )
      console.log(`state changed`)
    })
  }, [])
  return firebaseAuthStatus
}
