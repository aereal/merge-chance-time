import { createRouter, defineRoute } from "type-route"

const router = createRouter({
  root: defineRoute("/"),
  signIn: defineRoute("/sign-in"),
  showArticle: defineRoute({ id: "path.param.string" }, (p) => `/articles/${p.id}`),
})
export const { getCurrentRoute, routes, listen } = router
