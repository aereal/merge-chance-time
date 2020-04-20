export interface User {
  readonly token: string
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
