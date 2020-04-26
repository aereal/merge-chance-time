import { useState } from "react"

export const useFetchState = () => {
  const [fetchState, setFetchState] = useState<FetchState>({ kind: "ready" })
  return {
    fetchState,
    ready: () => {
      setFetchState(ready)
    },
    start: () => {
      setFetchState(started)
    },
    succeed: () => {
      setFetchState(successful)
    },
    fail: (error: Error) => {
      setFetchState({ kind: "failure", error })
    },
  }
}

const ready = { kind: "ready" } as const
export type Ready = typeof ready

const started = { kind: "started" } as const
export type Started = typeof started

export interface Failure {
  readonly kind: "failure"
  readonly error: Error
}

const successful = { kind: "successful" } as const
export type Successful = typeof successful

export type FetchState = Ready | Started | Successful | Failure
