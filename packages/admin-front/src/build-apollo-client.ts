import { ApolloClient } from "apollo-client"
import { HttpLink } from "apollo-link-http"
import { InMemoryCache, NormalizedCacheObject } from "apollo-cache-inmemory"
import { apiOrigin } from "./api-origin"

export const buildApolloClient = (token?: string): ApolloClient<NormalizedCacheObject> =>
  new ApolloClient({
    link: new HttpLink({
      uri: `${apiOrigin()}/api/query`,
      credentials: "include",
      headers: {
        ...(token !== undefined
          ? {
              authorization: `Bearer ${token}`,
            }
          : {}),
      },
      fetchOptions: {
        mode: "cors",
      },
    }),
    cache: new InMemoryCache(),
  })
