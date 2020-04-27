import React, { FC } from "react"
import { Route } from "type-route"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import LinearProgress from "@material-ui/core/LinearProgress"
import gql from "graphql-tag"
import { useQuery } from "@apollo/react-hooks"
import { routes } from "../routes"
import { GetRepoDetail, GetRepoDetailVariables } from "./__generated__/GetRepoDetail"
import { RepoDetail, REPO_DETAIL_FRAGMENT } from "../components/RepoDetail"

const GET_REPO_DETAIL = gql`
  query GetRepoDetail($owner: String!, $name: String!) {
    repository(owner: $owner, name: $name) {
      ...RepoDetailFragment
    }
  }
  ${REPO_DETAIL_FRAGMENT}
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
  return (
    <Grid item xs={12}>
      {loading && <LinearProgress />}
      {error !== undefined && <>Error: {JSON.stringify(error)}</>}
      {data && <RepoDetailPageContent {...data} />}
    </Grid>
  )
}

const RepoDetailPageContent: FC<GetRepoDetail> = ({ repository }) => {
  return (
    <>
      {repository ? (
        <>
          <Typography variant="subtitle1">
            {repository.owner.login}/{repository.name}
          </Typography>
          <RepoDetail repo={repository} />
        </>
      ) : (
        <>Not Found</>
      )}
    </>
  )
}
