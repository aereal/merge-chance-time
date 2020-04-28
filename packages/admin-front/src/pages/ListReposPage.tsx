import React, { FC } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import makeStyles from "@material-ui/core/styles/makeStyles"
import LinearProgress from "@material-ui/core/LinearProgress"
import gql from "graphql-tag"
import { useQuery } from "@apollo/react-hooks"
import { Helmet } from "react-helmet"
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
  const { loading, error, data } = useQuery<GetInstalledRepos>(GET_INSTALLED_REPOS)
  return (
    <>
      <Helmet>
        <title>Repos - Merge Chance Time</title>
      </Helmet>
      <Grid item xs={12}>
        {loading && <LinearProgress />}
        {error && <>Error: {JSON.stringify(error)}</>}
        {data && <ListReposPageContent {...data} />}
      </Grid>
    </>
  )
}

const ListReposPageContent: FC<GetInstalledRepos> = (data) => {
  const { root } = useStyles()
  return (
    <>
      <Typography variant="subtitle1">List Repos</Typography>
      <div className={root}>
        <ReposList repos={data.visitor.installations.flatMap((inst) => inst.installedRepositories)} />
      </div>
    </>
  )
}
