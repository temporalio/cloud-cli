package temporalcloudcli

func (c *CloudLoginCommand) run(cctx *CommandContext, args []string) error {
	_, err := login(cctx, nil)
	return err
}
