let mockGetItemResult: any
jest.mock("../local-storage", () => {
  return {
    getItem: () => {
      return mockGetItemResult
    },
  }
})

import { renderHook } from "@testing-library/react-hooks"
import { isUnauthenticated, isSignedIn } from "../auth"
import { useAuthentication } from "./authentication"

describe("useAuthentication", () => {
  test("ok", () => {
    mockGetItemResult = null
    const {
      result: {
        current: [authStatus],
      },
    } = renderHook(() => useAuthentication())

    expect(isUnauthenticated(authStatus)).toBe(true)
  })

  test("no token found", () => {
    mockGetItemResult = { token: "poppoe" }
    const {
      result: {
        current: [authStatus],
      },
    } = renderHook(() => useAuthentication())

    expect(isSignedIn(authStatus)).toBe(true)
  })
})
