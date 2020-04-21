import React, { FC, useState, useEffect } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import makeStyles from "@material-ui/core/styles/makeStyles"
import { ReposList } from "../components/ReposList"
import { Repo } from "../components/RepoSummary"
import { useAuthentication } from "../effects/authentication"
import { apiOrigin } from "../api-origin"
import { isSignedIn } from "../auth"

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
    backgroundColor: theme.palette.background.paper,
  },
}))

export const ListReposPage: FC = () => {
  const [authStatus] = useAuthentication()
  const [repos, setRepos] = useState<Repo[]>([])
  useEffect(() => {
    if (!isSignedIn(authStatus)) {
      return
    }

    const fetchData = async () => {
      const resp = await fetch(`${apiOrigin()}/api/user/installed_repos`, {
        headers: {
          authorization: `Bearer ${authStatus.user.token}`,
        },
        mode: "cors",
        credentials: "include",
      })
      const payload = await resp.json()
      setRepos(payload.repositories.map((r: any) => ({ fullName: r.full_name, owner: r.owner.login, name: r.name })))
    }
    fetchData()
  }, [authStatus.type])
  const { root } = useStyles()
  return (
    <Grid item xs={12}>
      <Typography variant="subtitle1">List Repos</Typography>
      <div className={root}>
        <ReposList repos={repos} />
      </div>
    </Grid>
  )
}
