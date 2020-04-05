import React, { FC, useState, useEffect } from "react"
import { Route } from "type-route"
import { routes, getCurrentRoute, listen } from "./routes"
import { RootPage } from "./pages/RootPage"

interface RoutingProps {
  readonly route: Route<typeof routes>
}

const Routing: FC<RoutingProps> = ({ route }) => {
  switch (route.name) {
    case routes.root.name:
      return <RootPage />
    default:
      return <>Not Found</>
  }
}

const App: FC = () => {
  const [route, setRoute] = useState(getCurrentRoute())
  useEffect(() => listen(setRoute), [])

  return <Routing route={route} />
}

export default App
