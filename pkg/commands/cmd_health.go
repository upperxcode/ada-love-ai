package commands

import "context"

func healthCommand() Definition {
	return Definition{
		Name:        "health",
		Description: "Check workspace health and configuration",
		Usage:       "/health",
		Handler: func(_ context.Context, req Request, rt *Runtime) error {
			if rt == nil || rt.WorkspaceHealth == nil {
				return req.Reply(unavailableMsg)
			}
			report, err := rt.WorkspaceHealth()
			if err != nil {
				return req.Reply("Health check failed: " + err.Error())
			}
			return req.Reply(report)
		},
	}
}