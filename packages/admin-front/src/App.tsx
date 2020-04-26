import React, { FC, useState, useEffect } from "react"
import CssBaseline from "@material-ui/core/CssBaseline"
import Container from "@material-ui/core/Container"
import Grid from "@material-ui/core/Grid"
import { Route } from "type-route"
import { DefaultAuthenticationProvider } from "./effects/authentication"
import { DefaultApolloClientProvider } from "./providers/apollo-client"
import { routes, getCurrentRoute, listen } from "./routes"
import { SignInPage } from "./pages/SignInPage"
import { CallbackPage } from "./pages/CallbackPage"
import { ListReposPage } from "./pages/ListReposPage"
import { RepoDetailPage } from "./pages/RepoDetailPage"
import { TokenPage } from "./pages/TokenPage"

interface RoutingProps {
  readonly route: Route<typeof routes>
}

const Routing: FC<RoutingProps> = ({ route }) => {
  switch (route.name) {
    case routes.signIn.name:
      return <SignInPage />
    case routes.root.name:
      return <ListReposPage />
    case routes.authCallback.name:
      return <CallbackPage />
    case routes.repoDetail.name:
      return <RepoDetailPage params={route.params} />
    case routes.token.name:
      return <TokenPage />
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
            <DefaultApolloClientProvider>
              <Routing route={route} />
            </DefaultApolloClientProvider>
          </DefaultAuthenticationProvider>
        </Grid>
      </Container>
    </>
  )
}

export default App
