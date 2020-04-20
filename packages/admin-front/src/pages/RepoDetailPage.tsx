import React, { FC } from "react"
import { Route } from "type-route"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import { routes } from "../routes"

interface RepoDetailPageProps {
  readonly params: Route<typeof routes.repoDetail>["params"]
}

export const RepoDetailPage: FC<RepoDetailPageProps> = ({ params }) => (
  <Grid item xs={12}>
    <Typography variant="subtitle1">
      {params.owner}/{params.name}
    </Typography>
  </Grid>
)
