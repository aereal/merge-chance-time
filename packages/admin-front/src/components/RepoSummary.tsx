import React, { FC } from "react"
import ListItem from "@material-ui/core/ListItem"
import ListItemText from "@material-ui/core/ListItemText"

export interface Repo {
  readonly fullName: string
}

interface RepoSummaryProps {
  readonly repo: Repo
}

export const RepoSummary: FC<RepoSummaryProps> = ({ repo }) => (
  <ListItem>
    <ListItemText primary={repo.fullName} />
  </ListItem>
)
