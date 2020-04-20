import React, { FC } from "react"
import ListItem from "@material-ui/core/ListItem"
import ListItemText from "@material-ui/core/ListItemText"
import { routes } from "../routes"

export interface Repo {
  readonly fullName: string
  readonly owner: string
  readonly name: string
}

interface RepoSummaryProps {
  readonly repo: Repo
}

export const RepoSummary: FC<RepoSummaryProps> = ({ repo }) => (
  <ListItem button {...routes.repoDetail.link({ owner: repo.owner, name: repo.name })}>
    <ListItemText primary={repo.fullName} />
  </ListItem>
)
