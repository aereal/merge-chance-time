import { createRouter, defineRoute } from "type-route"

const router = createRouter({
  root: defineRoute("/"),
  signIn: defineRoute("/sign-in"),
  firebaesAuth: defineRoute("/firebase-sign-in"),
  authCallback: defineRoute({ accessToken: "query.param.string.optional" }, () => "/auth/callback"),
  repoDetail: defineRoute(
    {
      owner: "path.param.string",
      name: "path.param.string",
    },
    ({ owner, name }) => `/repos/${owner}/${name}`
  ),
  token: defineRoute("/token"),
})
export const { getCurrentRoute, routes, listen } = router
