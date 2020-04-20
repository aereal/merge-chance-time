import { createRouter, defineRoute } from "type-route"

const router = createRouter({
  root: defineRoute("/"),
  signIn: defineRoute("/sign-in"),
  authCallback: defineRoute({ accessToken: "query.param.string.optional" }, () => "/auth/callback"),
  listRepos: defineRoute("/repos"),
  repoDetail: defineRoute(
    {
      owner: "path.param.string",
      name: "path.param.string",
    },
    ({ owner, name }) => `/repos/${owner}/${name}`
  ),
})
export const { getCurrentRoute, routes, listen } = router
