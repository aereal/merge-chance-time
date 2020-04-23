import React, { FC } from "react"
import { ApolloProvider } from "@apollo/react-hooks"
import { buildApolloClient } from "../build-apollo-client"
import { useAuthentication } from "../effects/authentication"
import { isSignedIn } from "../auth"

export const DefaultApolloClientProvider: FC = ({ children }) => {
  const [authStatus] = useAuthentication()
  const token = isSignedIn(authStatus) ? authStatus.user.token : undefined
  const client = buildApolloClient(token)
  return <ApolloProvider client={client}>{children}</ApolloProvider>
}
