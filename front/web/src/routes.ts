import { createRouter, defineRoute, Route } from "type-route"

const router = createRouter({
  root: defineRoute("/"),
})
export const { getCurrentRoute, routes, listen } = router
