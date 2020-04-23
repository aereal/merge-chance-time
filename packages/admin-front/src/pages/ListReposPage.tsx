import React, { FC } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import makeStyles from "@material-ui/core/styles/makeStyles"
import gql from "graphql-tag"
import { useQuery } from "@apollo/react-hooks"
import { ReposList } from "../components/ReposList"
import { REPO_SUMMARY } from "../components/RepoSummary"
import { GetInstalledRepos } from "./__generated__/GetInstalledRepos"

const GET_INSTALLED_REPOS = gql`
  query GetInstalledRepos {
    visitor {
      installations {
        installedRepositories {
          ...RepoSummary
        }
      }
    }
  }
  ${REPO_SUMMARY}
`

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
    backgroundColor: theme.palette.background.paper,
  },
}))

export const ListReposPage: FC = () => {
  const { root } = useStyles()
  const { loading, error, data } = useQuery<GetInstalledRepos>(GET_INSTALLED_REPOS)
  if (loading) {
    return <>Loading ...</>
  }
  if (error) {
    return <>Error: {JSON.stringify(error)}</>
  }
  if (!data) {
    return null
  }
  const repos = data.visitor.installations.flatMap((inst) => inst.installedRepositories)
  return (
    <Grid item xs={12}>
      <Typography variant="subtitle1">List Repos</Typography>
      <div className={root}>
        <ReposList repos={repos} />
      </div>
    </Grid>
  )
}
