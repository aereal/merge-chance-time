import { createRouter, defineRoute } from "type-route"

const router = createRouter({
  root: defineRoute("/"),
  signIn: defineRoute("/sign-in"),
  authCallback: defineRoute({ accessToken: "query.param.string.optional" }, () => "/auth/callback"),
})
export const { getCurrentRoute, routes, listen } = router