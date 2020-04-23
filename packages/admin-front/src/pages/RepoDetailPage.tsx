import React, { FC } from "react"
import { Route } from "type-route"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import gql from "graphql-tag"
import { useQuery } from "@apollo/react-hooks"
import { routes } from "../routes"
import { GetRepoDetail, GetRepoDetailVariables } from "./__generated__/GetRepoDetail"

const GET_REPO_DETAIL = gql`
  query GetRepoDetail($owner: String!, $name: String!) {
    repository(owner: $owner, name: $name) {
      owner {
        login
      }
      name
      config {
        startSchedule
        stopSchedule
        mergeAvailable
      }
    }
  }
`

interface RepoDetailPageProps {
  readonly params: Route<typeof routes.repoDetail>["params"]
}

export const RepoDetailPage: FC<RepoDetailPageProps> = ({ params }) => {
  const { loading, error, data } = useQuery<GetRepoDetail, GetRepoDetailVariables>(GET_REPO_DETAIL, {
    variables: {
      owner: params.owner,
      name: params.name,
    },
  })
  if (loading) {
    return <>Loading ...</>
  }
  if (error) {
    return <>Error: {JSON.stringify(error)}</>
  }
  if (!data || !data.repository) {
    return null
  }
  return (
    <Grid item xs={12}>
      <Typography variant="subtitle1">
        {data.repository.owner.login}/{data.repository.name}
      </Typography>
      <pre>{JSON.stringify(data.repository.config)}</pre>
    </Grid>
  )
}
