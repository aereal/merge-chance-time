import React, { FC, useState, useEffect } from "react"
import CssBaseline from "@material-ui/core/CssBaseline"
import Container from "@material-ui/core/Container"
import Grid from "@material-ui/core/Grid"
import { Route } from "type-route"
import { routes, getCurrentRoute, listen } from "./routes"
import { RootPage } from "./pages/RootPage"
import { SignInPage } from "./pages/SignInPage"
import { CallbackPage } from "./pages/CallbackPage"
import { ListReposPage } from "./pages/ListReposPage"
import { DefaultAuthenticationProvider } from "./effects/authentication"

interface RoutingProps {
  readonly route: Route<typeof routes>
}

const Routing: FC<RoutingProps> = ({ route }) => {
  switch (route.name) {
    case routes.signIn.name:
      return <SignInPage />
    case routes.root.name:
      return <RootPage />
    case routes.authCallback.name:
      return <CallbackPage />
    case routes.listRepos.name:
      return <ListReposPage />
    default:
      return <>Not Found</>
  }
}

const App: FC = () => {
  const [route, setRoute] = useState(getCurrentRoute())
  useEffect(() => listen(setRoute), [])

  return (
    <>
      <CssBaseline />
      <Container maxWidth="md">
        <Grid container>
          <DefaultAuthenticationProvider>
            <Routing route={route} />
          </DefaultAuthenticationProvider>
        </Grid>
      </Container>
    </>
  )
}

export default App
