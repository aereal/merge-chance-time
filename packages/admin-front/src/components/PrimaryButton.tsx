import React, { FC } from "react"
import Button, { ButtonProps } from "@material-ui/core/Button"

type PrimaryButtonProps = Omit<ButtonProps, "color" | "variant">

export const PrimaryButton: FC<PrimaryButtonProps> = ({ children, ...rest }) => (
  <Button color="primary" variant="contained" {...rest}>
    {children}
  </Button>
)
