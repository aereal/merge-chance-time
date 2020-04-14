import React, { FC, useState, useEffect } from "react"
import CssBaseline from "@material-ui/core/CssBaseline"
import Container from "@material-ui/core/Container"
import Grid from "@material-ui/core/Grid"
import { Route } from "type-route"
import { routes, getCurrentRoute, listen } from "./routes"
import { AuthenticationProvider } from "./effects/authentication"
import { RootPage } from "./pages/RootPage"
import { SignInPage } from "./pages/SignInPage"
import { CallbackPage } from "./pages/CallbackPage"

interface RoutingProps {
  readonly route: Route<typeof routes>
}

const Routing: FC<RoutingProps> = ({ route }) => {
  switch (route.name) {
    case routes.signIn.name:
      return <SignInPage />
    case routes.callback.name:
      return <CallbackPage />
    case routes.root.name:
      return <RootPage />
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
          <AuthenticationProvider>
            <Routing route={route} />
          </AuthenticationProvider>
        </Grid>
      </Container>
    </>
  )
}

export default App
