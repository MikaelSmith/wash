package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	apitypes "github.com/puppetlabs/wash/api/types"
	"github.com/puppetlabs/wash/cmd/internal/config"
	cmdutil "github.com/puppetlabs/wash/cmd/util"
	"github.com/puppetlabs/wash/plugin"
)

func listCommand() *cobra.Command {
	aliases := []string{}
	if !config.Embedded {
		aliases = []string{"ls"}
	}
	listCmd := &cobra.Command{
		Use:     "list [<file>]",
		Aliases: aliases,
		Short:   "Lists the resources at the indicated path",
		Args:    cobra.MaximumNArgs(1),
		RunE:    toRunE(listMain),
	}
	return listCmd
}

func headers() []cmdutil.ColumnHeader {
	return []cmdutil.ColumnHeader{
		{ShortName: "name", FullName: "NAME"},
		{ShortName: "mtime", FullName: "MODIFIED"},
		{ShortName: "verbs", FullName: "ACTIONS"},
	}
}

func format(t time.Time) string {
	return t.Format(time.RFC822)
}

func formatListEntries(ls []apitypes.Entry) string {
	table := make([][]string, len(ls))
	for i, entry := range ls {
		var mtimeStr string
		if entry.Attributes.HasMtime() {
			mtimeStr = format(entry.Attributes.Mtime())
		} else {
			mtimeStr = "<unknown>"
		}

		verbs := strings.Join(entry.Actions, ", ")

		name := entry.CName
		if len(ls) > 1 && i == 0 {
			// Represent the pwd as "."
			name = "."
		}

		if entry.Supports(plugin.ListAction()) {
			name += "/"
		}

		table[i] = []string{name, mtimeStr, verbs}
	}
	return cmdutil.NewTableWithHeaders(headers(), table).Format()
}

func listMain(cmd *cobra.Command, args []string) exitCode {
	// If no path is declared, try to list the current directory/resource
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	conn := cmdutil.NewClient()
	e, err := conn.Info(path)
	if err != nil {
		cmdutil.ErrPrintf("%v\n", err)
		return exitCode{1}
	}
	entries := []apitypes.Entry{e}
	if e.Supports(plugin.ListAction()) {
		children, err := conn.List(path)
		if err != nil {
			cmdutil.ErrPrintf("%v\n", err)
			return exitCode{1}
		}
		entries = append(entries, children...)
	}

	cmdutil.Print(formatListEntries(entries))
	return exitCode{0}
}
